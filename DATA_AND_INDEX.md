# 数据与索引

以数据主要存放的地点来划分，数据库可以分为 in-memory 和 disk-based 两种。前者将所有数据放进内存中，读写的过程不存在磁盘 I/O；后者将所有数据存放在磁盘中，读写过程需要频繁与磁盘交互，Bolt 就属于后者。对于 disk-based 数据库来说，数据库的性能瓶颈常常出现在磁盘 I/O 上，因此磁盘 I/O 次数常常被用来衡量这类数据库各操作的性能，减少磁盘 I/O 也就成为重要的设计目标。B+ 树正是为优化数据查询发生的磁盘 I/O 而生，它可以很好地将数据查询与块存储相结合，减少每次数据读取所需的 I/O 次数。Bolt 使用 B+ 树作为索引，将所有键值数据放在叶子节点。

磁盘中的一个 page 对应 B+ 树上的一个 node，数据库从磁盘载入 page 的过程就是将其反序列化成 node 的过程，将数据写入磁盘的过程就是将 node 序列化成 page 的过程。在[存储与缓存](STORAGE_AND_CACHE.md)一节中介绍的 branch page 和 leaf page 对应的正是 B+ 树上的 branch node 和 leaf node。Bolt 是键值数据库，不是关系型数据库，可以理解成每条数据只有一个字段，因此也只需要在这个字段上建立索引。Bolt 直接将数据存储在 B+ 树的 leaf node 上，类似关系型数据库中的一级索引。

为了方便，本文将以 node 为中心来讨论 Bolt 中的 B+ 树的结构及相关算法。

## 目录

* 结构
* 算法
  * 插入和删除
  * rebalance 与 spill
* 小结

## 结构

B+ 树上的每个 node 的结构都由 header，element header 列表及数据列表依次构成，branch node 中存储的数据只有键，leaf node 中存储的数据则是键值对：

![branch/leaf page](./statics/imgs/data-and-index-branch-leaf-page.jpg)

简化后的 branch node 如下：

![branch page/node](./statics/imgs/data-and-index-branch-page-node.jpg)

简化后的 leaf node 如下：

![leaf page/node](./statics/imgs/data-and-index-leaf-page-node.jpg)

利用简化后的 branch node 和 leaf node，就能够构建出一棵合法的 B+ 树：

![b+tree](./statics/imgs/data-and-index-b-plus-tree.jpg)

这棵树的特性基本与教科书上的类似，但略有不同：

* 整棵 B+ 树是完美平衡树，即每个 leaf node 的深度都一样;
* 除 root node之外，其它 node 的填充率需要大于 FillPercent，默认为 50%;
* 所有的键值对都存储在 leaf node 中，branch node 只存储每个 child node 的最小键 (如上图箭头所示)，键的大小由键的字节序决定。

## 算法

与教科书上介绍的 B+ 树操作不同，Bolt 在更新 B+ 树数据时不会直接修改树的结构，而是只更新数据。在数据写入磁盘前才按需合并、分裂 node，恢复 B+ 树的不变性；严格地说，Bolt 中的 B+ 树是一棵磁盘中的 B+ 树，在内存中执行更新操作且尚未写入磁盘前，它可能不符合 B+ 树的特性。

### 插入和删除

由于 Bolt 在更新 B+ 树时不会直接修改树的结构，只更新数据，读写事务插入数据或删除数据只需执行两步：

* 顺着 root node 找到目标键所在的 leaf node
* 插入或删除键值对数据

*插入*

插入数据可能造成 node 中的数据超过单个 page 大小：

![oversize-page](./statics/imgs/data-and-index-oversize-page.jpg)

同时也要考虑单个键值对数据过大的情况：

![oversized-kv](./statics/imgs/data-and-index-oversized-kv.jpg)

*删除*

删除数据可能使得 node 处于不平衡的状态，即键值对删除后 node 中的总数据量少于最低填充率 ( FillPercent)：

![unbalanced-node](./statics/imgs/data-and-index-unbalanced-node.jpg)

当删除操作导致 node 的填充率低于要求时，虽然不会立即与兄弟节点合并，但会在结束前将该 node 标记为 `unbalanced`，等数据要写入磁盘前再统一合并。

### rebalance 与 spill

在读写事务提交时，Bolt 通过 rebalance 和 spill 来保证最终写入磁盘的是一棵合法的 B+ 树：

* rebalance：将填充率不足的 node 与 sibling node 合并
* spill：将填充率过高的 node 分裂成多个较小的 node

*rebalance*

> rebalance attempts to combine the node with sibling nodes if the node fill size is below a threshold.

执行过删除操作的 leaf node 都会被打上 unbalanced 的记号，而这些记号的消费者正是 rebalance，它负责将填充率不足 `FillPercent` 的 node 与对应的 sibling node 合并，以下图为例 (图中虚线为填充率)：

![rebalance-1](./statics/imgs/data-and-index-rebalance-1.jpg)

当 L2 的数据总量不足，rebalance 找到它的 sibling node L3，将二者合并，并更新它们的 parent node B2：

![rebalance-2](./statics/imgs/data-and-index-rebalance-2.jpg)

这时，B2 的数据总量小于 `FillPercent`，rebalance 需要递归地继续合并 B2 与 B3：

![rebalance-3](./statics/imgs/data-and-index-rebalance-3.jpg)

如果有必要，这样的递归过程将持续到 root node 为止。这里还有一种特殊情况需要考虑：当 root node 为 branch node，且它只有一个 child node 时，可以直接将唯一的 child node 提拔成 root node，并释放原先的 root node。

rebalance 之后，内存中的树装结构满足：所有 node 的数据填充率都大于 `FillPercent`。但这些 node 的填充率可能远远大于 100%，因此就需要 spill 操作从反方向将所有 node 的填充率控制在 `FillPercent` 之下。

*spill*

> spill writes the nodes to dirty pages and splits nodes as it goes.

*注：spill 除了将 node 拆分，还会将其转化成 page，以下讨论不涉及转化的部分。*

spill 本意是「水太满而从容器中溢出」，这里指的就是 node 中填充的数据太多，溢出到多个 node，举例如下：

![spill-1](./statics/imgs/data-and-index-spill-1.jpg)

显然 L1 中填充率过高，spill 需要将它拆分成多个 node：

![spill-2](./statics/imgs/data-and-index-spill-2.jpg)

拆分 L1 的过程会导致它的 parent node 填充率升高，进而可能引发 parent node 的拆分：

![spill-3](./statics/imgs/data-and-index-spill-3.jpg)

如果有必要，这样的递归会到达 root node 为止。

parent node 超载的原因除了键值对数量过多，也可能是单个数据过大。如果是后面这种情况，spill 不会再将 node 拆分，而是保留这些超载的 node：

![spill-4](./statics/imgs/data-and-index-spill-4.jpg)

在序列化时，这种超载的 node 会被转化为 overflow page 存储。

## 小结

Bolt 中的 B+ 树实现：

* 将键值数据直接存储在叶子节点
* 支持存储变长键值数据
* 在内存中允许 B+ 树发生临时「变形」，落盘前再统一矫正，保证磁盘中的 B+ 树符合要求

## 参考

* [boltdb src code: node.go](https://github.com/boltdb/bolt/blob/master/node.go)
* [Lessons Learned While Implementing a B+Tree](https://hackthology.com/lessons-learned-while-implementing-a-btree.html)
