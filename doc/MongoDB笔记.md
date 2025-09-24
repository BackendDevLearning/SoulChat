```

```

# MongoDB



## 基本特点

文档型数据库，数据以类似 **JSON 的文档形式存储**。

MongoDB 的设计理念是为了应对 **大数据量**、高性能和 **灵活性** 需求。

MongoDB使用集合（Collections）来组织文档（Documents），每个**文档都是由键值对**组成的。

------

MongoDB 将数据存储为一个文档，数据结构由键值(key=>value)对组成，文档类似于 JSON 对象，字段值可以包含其他文档，数组及文档数组

- **数据库（Database）**：存储数据的容器，类似于关系型数据库中的**数据库**。
- **集合（Collection）**：数据库中的一个集合，类似于关系型数据库中的**表**。
- **文档（Document）**：集合中的一个**数据记录的基本单元**，类似于关系型数据库中的行（row），以 BSON 格式存储。

![img](https://www.runoob.com/wp-content/uploads/2013/10/Figure-1-Mapping-Table-to-Collection-1.png)

## 基本概念

| SQL 术语/概念 | MongoDB 术语/概念 | 解释/说明                                 |
| :------------ | :---------------- | :---------------------------------------- |
| database      | database          | 数据库                                    |
| table         | **collection**    | 数据库表/集合                             |
| row           | **document**      | 数据记录行/文档                           |
| column        | **field**         | 数据字段/域                               |
| index         | index             | 索引                                      |
| table joins   |                   | 表连接,MongoDB不支持                      |
| primary key   | primary key       | 主键,MongoDB自动**将 _id 字段设置为主键** |



**完整术语列表：**

- **文档（Document）**：MongoDB 的基本数据单元，通常是一个 JSON-like 的结构，可以**包含多种数据类型**。
- **集合（Collection）**：类似于关系型数据库中的**表**，集合是一组文档的容器。在 MongoDB 中，**一个集合中的文档不需要有一个固定的模式**。
- **数据库（Database）**：包含一个或多个集合的 MongoDB 实例。
- **BSON**：Binary JSON 的缩写，是 MongoDB 用来存储和传输文档的 **二进制形式的 JSON**。
- **索引（Index）**：用于优化查询性能的数据结构，可以基于集合中的**一个或多个**字段创建索引。
- **分片（Sharding）**：一种分布数据到多个服务器（称为分片）的方法，用于处理大数据集和高吞吐量应用。
- **副本集（Replica Set）**：一组维护相同数据集的 MongoDB 服务器，提供数据的冗余备份和高可用性。
- **主节点（Primary）**：副本集中负责处理所有**写入**操作的服务器。
- **从节点（Secondary）**：副本集中的服务器，用于**读取**数据和在主节点故障时接管为主节点。
- **MongoDB Shell**：MongoDB 提供的命令行界面，用于与 MongoDB 实例交互。
- **聚合框架（Aggregation Framework）**：用于执行复杂的数据处理和聚合操作的一系列操作。
- **Map-Reduce**：一种编程模型，用于处理大量数据集的并行计算。
- **GridFS**：用于存储和检索大于 BSON 文档大小限制的文件的规范。
- **ObjectId**：MongoDB 为每个文档自动生成的唯一标识符。
- **CRUD 操作**：创建（Create）、读取（Read）、更新（Update）、删除（Delete）操作。
- **事务（Transactions）**：从 MongoDB 4.0 开始支持，允许一组操作作为一个原子单元执行。
- **操作符（Operators）**：用于查询和更新文档的特殊字段。
- **连接（Join）**：MongoDB 允许在查询中使用 `$lookup` 操作符来实现类似 SQL 的连接操作。
- **TTL（Time-To-Live）**：可以为集合中的某些字段设置 TTL，以自动删除旧数据。
- **存储引擎（Storage Engine）**：MongoDB 用于数据存储和管理的底层技术，如 WiredTiger 和 MongoDB 的旧存储引擎 MMAPv1。
- **MongoDB Compass**：MongoDB 的图形界面工具，用于可视化和管理 MongoDB 数据。
- **MongoDB Atlas**：MongoDB 提供的云服务，允许在云中托管 MongoDB 数据库。



## 文档(Document)

文档是**一组键值(key-value)对(即 BSON)**。MongoDB 的文档**不需要设置相同的字段**，并且**相同的字段不需要相同的数据类型**，这与关系型数据库有很大的区别，也是 MongoDB 非常突出的特点。

一个简单的文档例子如下：

```
{"site":"www.runoob.com", "name":"菜鸟教程"}
```



#### 需要注意的是：

1. 文档中的 **键/值对是有序** 的。
2. 文档中的值不仅可以是在双引号里面的字符串，还可以是其他几种数据类型（甚至可以是整个嵌入的文档)。
3. MongoDB **区分类型** 和 **大小写**。
4. MongoDB的文档**不能有重复的键**。
5. 文档的键是字符串。除了少数例外情况，键可以使用任意UTF-8字符。

#### 文档键命名规范：

- \0 (**空字符**) 用来表示 **键的结尾**，键的前面不能有这个。
- .和$有特别的意义，只有在特定环境下才能使用。
- 以下划线**"_"开头的键**是保留的(不是严格要求的)。





## 集合（collection）

集合就是 MongoDB 文档组，其实就是sql的表

集合存在于数据库中，集合没有固定的结构，这意味着你在对集合可以插入不同格式和类型的数据，但通常情况下我们插入集合的数据都会有一定的关联性。

比如，我们可以将以下不同数据结构的文档插入到集合中：

```
{"site":"www.baidu.com"}
{"site":"www.google.com","name":"Google"}
{"site":"www.runoob.com","name":"task","num":5}
```

当第一个文档插入时，集合就会被创建

### 合法的集合名

- 集合名不能是空字符串""。
- 集合名不能含有**\0字符（空字符)**，这个字符表示集合名的结尾。
- 集合名不能以**"system."**开头，这是为系统集合保留的前缀。
- 用户创建的集合名字不能含有保留字符。有些驱动程序的确支持在集合名里面包含，这是因为某些系统生成的集合中包含该字符。除非你要访问这种系统创建的集合，否则千万不要在名字里出现**$**。　





### capped collections

Capped collections 就是 **固定大小** 的collection。

它有很高的性能以及 **队列过期** 的特性(过期按照插入的顺序). 有点和 "RRD" 概念类似。

Capped collections 是 **高性能自动的维护对象的插入顺序**。它非常适合类似记录日志的功能和标准的 collection 不同，你**必须要显式的创建一个capped collection**，**指定**一个 collection 的**大小**，单位是**字节**。collection 的数据存储空间值**提前分配**的。

Capped collections 可以按照文档的插入顺序保存到集合中，而且这些文档在**磁盘上存放位置也是按照插入顺序来保存**的，所以当我们更新Capped collections 中文档的时候，更新后的文档不可以超过之前文档的大小，这样话就可以确保所有文档在磁盘上的位置一直保持不变。

由于 Capped collection 是按照文档的**插入顺序而不是使用索引确定插入位置**，这样的话可以提**高增添数据的效率**。MongoDB 的操作日志文件 oplog.rs 就是利用 Capped Collection 来实现的。

要注意的是指定的存储大小包含了数据库的头信息。



```
db.createCollection("mycoll", {capped:true, size:100000})
```

- 在 capped collection 中，你能添加新的对象。
- 能进行更新，然而，对象不会增加存储空间。如果增加，更新就会失败 。
- 使用 Capped Collection 不能删除一个文档，可以使用 drop() 方法删除 collection 所有的行。
- 删除之后，你必须显式的重新创建这个 collection。
- 在32bit机器中，capped collection 最大存储为 1e9( 1X109)个字节。





## MongoDB 数据类型

下表为MongoDB中常用的几种数据类型。

| 数据类型           | 描述                                                         |
| :----------------- | :----------------------------------------------------------- |
| String             | 字符串。存储数据常用的数据类型。在 MongoDB 中，UTF-8 编码的字符串才是合法的。 |
| Integer            | 整型数值。用于存储数值。根据你所采用的服务器，可分为 32 位或 64 位。 |
| Boolean            | 布尔值。用于存储布尔值（真/假）。                            |
| Double             | 双精度浮点值。用于存储浮点值。                               |
| Min/Max keys       | 将一个值与 BSON（二进制的 JSON）元素的最低值和最高值相对比。 |
| Array              | 用于将数组或列表或多个值存储为一个键。                       |
| Timestamp          | 时间戳。记录文档修改或添加的具体时间。                       |
| Object             | 用于内嵌文档。                                               |
| Null               | 用于创建空值。                                               |
| Symbol             | 符号。该数据类型基本上等同于字符串类型，但不同的是，它一般用于采用特殊符号类型的语言。 |
| Date               | 日期时间。用 UNIX 时间格式来存储当前日期或时间。你可以指定自己的日期时间：创建 Date 对象，传入年月日信息。 |
| Object ID          | 对象 ID。用于创建文档的 ID。                                 |
| Binary Data        | 二进制数据。用于存储二进制数据。                             |
| Code               | 代码类型。用于在文档中存储 JavaScript 代码。                 |
| Regular expression | 正则表达式类型。用于存储正则表达式。                         |



### ObjectId

ObjectId 类似唯一主键，可以很快的去生成和排序，包含 12 bytes，含义是：

- 前 4 个字节表示创建 **unix** 时间戳,格林尼治时间 **UTC** 时间，比北京时间晚了 8 个小时
- 接下来的 3 个字节是机器标识码
- 紧接的两个字节由进程 id 组成 PID
- 最后三个字节是随机数


![img](https://www.runoob.com/wp-content/uploads/2013/10/2875754375-5a19268f0fd9b_articlex.jpeg)

MongoDB 中存储的文档必须有一个 _id 键。这个键的值可以是任何类型的，默认是个 ObjectId 对象

由于 ObjectId 中保存了创建的时间戳，所以你不需要为你的文档保存时间戳字段，你可以通过 getTimestamp 函数来获取文档的创建时间:

```
> var newObject = ObjectId()
> newObject.getTimestamp()
ISODate("2017-11-25T07:21:10Z")
```

ObjectId 转为字符串

```
> newObject.str
5a1919e63df83ce79df8b38f
```


