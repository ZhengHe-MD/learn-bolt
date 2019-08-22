# learn-boltdb

在最近的闲暇时间，我开始弥补自己数据库知识的盲点。前期花费若干个月时间认真阅读学习了：

* CMU 15-445/645 Intro to Database Systems
* Designing Data-Intensive Applications (DDIA)

站在巨人的肩膀上，可以体会到数据库世界的广袤，已然与落后几十年的教科书描述的那般景象大相径庭。但这些努力始终只是在走近，而不是走进数据库。因此我迫切地想要认真阅读某个数据库的源码，了解那些常被提及的概念是如何被实现。在 DDIA 中 boltDB 被多次提及，对它最深刻的印象就是通过单线程执行读写事务的方式简单粗暴地实现最苛刻的可序列化的事务隔离。于是有一天，我心血来潮访问了 boltDB 项目，发现它极简的设计理念正好满足我的学习需求。

## 为什么选择 boltDB

> The original goal of Bolt was to provide a simple pure Go key/value store and to not bloat the code with extraneous features.  --- Ben Johnson

boltDB 可能是最适熟悉 golang 的工程师最适合阅读的第一个数据库项目，原因在于它：

* 功能简单：
  * 它是键值数据库，没有关系的抽象，只有键和值
  * 它对外只暴露简单的 api，没有 DSL 层
  * 所有数据有且只有一个 B+ 树索引，没有查询计划
  * 同时允许多个只读事务和最多一个读写事务运行，有并发场景，有事务隔离，但处理方式极简
  * 使用 mmap 作读缓存，缓存管理逻辑达到最简
* 源码结构简单，从存储层到对外暴露的 api 抽象分层思路清晰：
  * 存储与缓存层：page.go, freelist.go, bolt_xxx.go
  * 数据与索引：node.go
  * 桶：bucket.go, cursor.go
  * 事务：tx.go
  * API：db.go
* 已经投入生产使用，功能稳定
* 项目已经定稿、归档，不再更新，就如一本已经刊印的书，可以放心引用、批注

## 关于本系列文章

*要达到好的学习效果，就要有输出*。以我平时的工作节奏，在闲暇时间依葫芦画瓢写一个键值数据库不太现实。于是我选择将自己对源码阅读心得系统地记录下来，最终整理成本系列文章，旨在尽我所能正确地描述 boltDB。恰好我在多次尝试在网上寻找相关内容后，发现网上大多数的文章、视频仅仅是介绍 boltDB 的用法和特性。因此，也许本系列文章可以作为它们以及 [boltDB 官方文档](https://github.com/boltdb/bolt/blob/master/README.md) 的补充，帮助想了解它的人更快地、深入地了解 boltDB。

如果你和我一样是初学者，相信它对你会有所帮助；如果你是一名经验丰富的数据库工程师，也许本系列文章对你来说没有太多新意。**欢迎所有读者通过任意方式反馈文章中的错误**，帮助我们共同进步。

### 阅读建议

每篇文章将从以某个抽象层为中心，描述 boltDB 的实现思路，它们按自底向上的顺序介绍 boltDB，但各个模块又相对独立，因此顺序阅读和单篇阅读皆可。除了系列文章之外，我在阅读过程中，也会直接在源码上做一些额外的注释帮助理解，如果你愿意，也可以 fork 我的 boltDB fork 来阅读，所有额外添加的注释格式如下：

```go
// [M]
// 中文版本
//   ...
// English version
//   ...
```

### 名词解释

为了减小歧义，涉及到专业术语的部分，除了已经有稳定的、确定的中文翻译之外，文章中会保留原始英文单词。下面是一份中英文对照表，供参考：

| 中文                  | 英文                                 |
| --------------------- | ------------------------------------ |
| B+树                  | B+Tree                               |
| 事务                  | transaction/tx                       |
| 读写事务              | read-write transaction/tx            |
| 只读事务              | read-only transaction/tx             |
| 桶                    | bucket                               |
| 游标                  | cursor                               |
| 键/值/键值对/键值数据 | key/value/key-value(kv) pair/kv data |
| 分裂                  | split                                |
| 合并                  | merge                                |

除此以外，一些名词需要特别提及：

#### 1. Page

物理外存和物理内存通常会被分割为同样大小的块，这些块被称为 page frame/frame，是物理概念。当这些块被读入内存，在程序中被引用、读写时，被称为 page，是逻辑概念。在 boltDB 中，二者都可能被称为 page，如:

* pageSize 中的 page 指的是物理概念

* page.go 中的大部分 page 指的是逻辑概念

#### 2. boltDB 与 boltDB 实例

boltDB 指这个数据库[项目](https://github.com/boltdb/bolt)本身，而 boltDB 实例指利用 api 创建的一个新的 database，也指在文件系统上相应创建的二进制数据库文件 "xxx.db"。

## 目录

本系列文章将按自底向上地顺序介绍 boltDB，目录如下：

* [存储与缓存](./STORAGE_AND_CACHE.md)
* [数据与索引](./DATA_AND_INDEX.md)
* [桶](./BUCKET.md)
* [事务](./TX.md)
* API [TODO]

## 其它

转载请注明出处为本项目: https://github.com/ZhengHe-MD/learn-bolt

## 参考

* Code
  * [Github: bolt/boltdb](https://github.com/boltdb/bolt)
* Video
  - [CMU 15-445/645 Intro to Database Systems](https://www.youtube.com/playlist?list=PLSE8ODhjZXja3hgmuwhf89qboV1kOxMx7)
* Book
  - [Designing Data-Intensive Applications](https://dataintensive.net/)

