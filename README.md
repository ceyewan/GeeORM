# GeeORM

对象关系映射（ORM，Object Relational Mapping）是一种编程技术，用于实现面向对象编程语言里不同类型系统的数据之间的转换。从效果上说，它其实是用面向对象思想，把数据库里的数据表示出来。ORM 框架旨在使开发人员能够通过对象操作完成数据库交互，而无需编写原生 SQL 语句。

| 数据库元素     | 面向对象元素         |
| --------- | -------------- |
| 表（Table）  | 类（Class）       |
| 记录（Row）   | 对象（Object）     |
| 字段（Field） | 对象属性（Property） |

以下示例展示了 SQL 语句与 ORM 框架代码的等效性，以辅助理解：

```go
-- SQL 语句
CREATE TABLE User (Name text, Age integer);  
INSERT INTO User (Name, Age) VALUES ("Tom", 18);  
SELECT * FROM User;

-- ORM 框架代码
type User struct {  
    Name string  
    Age  int  
}  
  
orm.CreateTable(&User{})  
orm.Save(&User{"Tom", 18})  
var users []User  
orm.Find(&users)
```

ORM 框架通过结构体与数据库表进行映射，避免了手动编写 SQL 语句的繁琐，并提高了跨数据库平台的兼容性。开发者只需关注对象的操作，即可实现数据的持久化。

ORM 框架的通用性依赖于 Go 语言的**反射机制**，用于动态获取结构体的元数据。反射允许程序在运行时检查和操作对象的类型、字段和方法等信息。

```go
func main() {
    user := User{Name: "Alice", Age: 25}
    // 获取 reflect.Value
    value := reflect.Indirect(reflect.ValueOf(&user)）
    fmt.Printf("Value: %v\n", value)
    // 获取 reflect.Type
    typ := value.Type()
    fmt.Printf("Type: %v\n", typ)
    // 遍历结构体字段
    for i := 0; i < value.NumField(); i++ {
        field := value.Field(i) // 获取字段值
        fieldType := typ.Field(i) // 获取字段类型信息
        // 打印字段名称、类型和值
        fmt.Printf("Field Name: %s, Field Type: %s, Field Value: %v\n",
        fieldType.Name, fieldType.Type, field.Interface())
    }
}
```

运行结果如下：

```text
Value: {Alice 25}
Type: main.User
Field Name: Name, Field Type: string, Field Value: Alice
Field Name: Age, Field Type: int, Field Value: 25
```

结果分析：

- `reflect.ValueOf()`: 获取对象的反射值。
- `reflect.Indirect()`: 获取指针指向的对象的反射值。
- `reflect.Type()`: 获取对象的类型信息，例如 `main.User`。
- `reflect.Type.Field(i)`: 获取第 i 个字段的信息，返回 `StructField` 类型，包含字段名称和类型等。

通过 Go 反射机制，开发人员可以动态地处理结构体，这使得 ORM 框架能够自动映射数据库表与对象之间的关系。反射使得在运行时访问类型的结构成为可能，进而动态地处理对象的属性。对于 ORM 框架而言，这种机制是其通用性和灵活性的基础。

## 1 database/sql 基础

### 1.1 初识 SQLite

SQLite 是一款轻量级的关系型数据库，它实现了 ACID 事务，并可以直接嵌入到应用程序中。与 MySQL、PostgreSQL 等数据库不同，SQLite 不需要独立的服务器进程，因此非常适合作为示例数据库。

### 1.2 database/sql 标准库

Go 语言提供了 `database/sql` 标准库，用于与数据库进行交互。

```go
func main() {
    db, _ := sql.Open("sqlite3", "gee.db")
    defer func() { _ = db.Close() }()
    _, _ = db.Exec("DROP TABLE IF EXISTS User;")
    _, _ = db.Exec("CREATE TABLE User(Name text);")
    result, err := db.Exec("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam")
    if err == nil {
        affected, _ := result.RowsAffected()
        log.Println(affected)
    }
    row := db.QueryRow("SELECT Name FROM User LIMIT 1")
    var name string
    if err := row.Scan(&name); err == nil {
        log.Println(name)
    }
}
```

使用 `sql.Open()` 连接数据库，第一个参数是驱动名称，第二个参数是数据库的名称，对于 SQLite 来说，也就是文件名，不存在会新建。返回一个 `sql.DB` 实例的指针。

`Exec()` 用于执行 SQL 语句，如果是查询语句，不会返回相关的记录。所以查询语句通常使用 `Query()` 和 `QueryRow()`，前者可以返回多条记录，后者只返回一条记录。

`Exec()`、`Query()`、`QueryRow()` 接受1或多个入参，第一个入参是 SQL 语句，后面的入参是 SQL 语句中的占位符 `?` 对应的值，占位符一般用来防 SQL 注入。

`QueryRow()` 的返回值类型是 `*sql.Row`，`row.Scan()` 接受1或多个指针作为参数，可以获取对应列(column)的值，在这个示例中，只有 `Name` 一列，因此传入字符串指针 `&name` 即可获取到查询的结果。

## 2 Clause

`clause` 模块的核心作用是构建 SQL 查询语句及其参数。它通过将 SQL 语句的各个组成部分（如 `SELECT`、`WHERE`、`LIMIT` 等）抽象为独立的组件，使得能够灵活地组合这些组件，从而生成完整的 SQL 语句。以下将结合代码详细解释 `clause` 模块的实现及其在 ORM 中的作用。

`Clause` 结构体是 `clause` 模块的核心，它负责存储 SQL 语句的各个组成部分及其对应的参数。结构体定义如下：

```go
type Clause struct {
 sql     map[Type]string        // 存储不同类型的 SQL 语句
 sqlVars map[Type][]interface{} // 存储 SQL 语句对应的参数
}
```

- `sql`：一个 `map`，键为 `Type` 类型（表示 SQL 语句的类型，如 `SELECT`、`WHERE` 等），值为对应的 SQL 语句片段。
- `sqlVars`：一个 `map`，键为 `Type` 类型，值为 SQL 语句片段对应的参数。

该模块主要包含两个函数：`Set` 方法用于设置特定类型的 SQL 语句及其对应的参数。它通过调用 `generators` 中的对应生成器函数来生成 SQL 语句片段和参数。`Build` 方法用于根据指定的顺序构建最终的 SQL 语句及其参数。它按照传入的 `orders` 顺序，将各个 SQL 片段拼接成完整的 SQL 语句，并将参数合并。例如，以下代码会调用两次 `Set`，最后调用一次 `Build` 生成最终的 SQL 语句。

```go
s.Where("Name = ?", "Tom").Update("Age", 30)
```

`generator.go` 文件中定义了一系列 `generator` 函数，用于生成不同类型的 SQL 语句片段。例如：

- `_select`：生成 `SELECT` 语句片段。
- `_where`：生成 `WHERE` 语句片段。
- `_limit`：生成 `LIMIT` 语句片段。

这些函数通过接收不同的参数，生成对应的 SQL 片段和参数列表。例如：

```go
func _select(values ...interface{}) (string, []interface{}) {
 tableName := values[0]
 fields := strings.Join(values[1].([]string), ", ")
 return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}
```

- `_select` 函数接收表名和字段列表，生成 `SELECT` 语句片段。

`clause` 模块通过将 SQL 语句的各个部分抽象为独立的组件，实现了 SQL 语句的灵活构建。`Clause` 结构体负责存储 SQL 片段及其参数，`Set` 方法用于设置 SQL 片段，`Build` 方法用于拼接完整的 SQL 语句。这种设计使得 ORM 框架能够轻松地生成复杂的 SQL 查询，同时保持代码的可维护性和扩展性。

通过 `clause` 模块，ORM 框架可以将面向对象的操作转换为底层的 SQL 语句，从而实现对象与数据库表之间的映射。这是 ORM 框架的核心功能之一。

## 3 Dialect

在实现 ORM（对象关系映射）时，**数据库方言（Dialect）** 是一个关键组件。它用于处理不同数据库之间的差异，使得 ORM 能够兼容多种数据库。

在 `dialect.go` 中，定义了一个 `Dialect` 接口，它包含两个核心方法：

- **DataTypeOf**: 将 Go 语言的类型映射到数据库中的数据类型。
- **TableExistSQL**: 生成检查表是否存在的 SQL 语句。

这个接口的作用是抽象出不同数据库的差异，使得 ORM 可以通过统一的接口处理多种数据库。

通过 `RegisterDialect` 和 `GetDialect` 实现了方言的注册和获取机制：

- **RegisterDialect**: 将方言实例注册到全局的 `dialectsMap` 中。
- **GetDialect**: 根据方言名称从 `dialectsMap` 中获取对应的方言实例。

这种机制使得 ORM 能够动态地支持多种数据库方言。

## 4 Schema

`schema` 模块的核心作用是实现对象（Go 结构体）与数据库表之间的映射关系。它负责解析 Go 结构体，生成对应的数据库表结构（`Schema`），并提供方法来操作这些结构。下面我们将结合代码详细解释 `schema` 模块的实现及其在 ORM 中的作用。

```go
// Field 表示数据库表的一列
type Field struct {
    Name string // 列名
    Type string // 列的数据类型
    Tag string // 列的额外信息（标签）
}

// Schema 表示数据库中的一张表
type Schema struct {
    Model interface{} // 表对应的对象
    Name string // 表名
    Fields []*Field // 表的所有列
    FieldNames []string // 表的所有列名
    fieldMap map[string]*Field // 列名到 Field 对象的映射
}
```

该模块主要包含三个函数：`GetField()` 可以根据列名获取 `Field` 对象；`Parse()` 解析 Go 结构体，创建 `Schema` 对象，这个方法主要是通过反射机制遍历结构体的字段生成表的列信息；`RecordValues()` 方法用于获取 Go 结构体对象中所有字段的值，通常用于将对象的值插入到数据库表中。

通过 `dialect` 接口，`schema` 模块可以支持不同类型的数据库（如 MySQL、PostgreSQL 等），并根据数据库方言生成正确的数据类型。

## 5 Session

Session 主要复制与数据库会话和 CRUD 操作。

```go
// Session 是会话管理的主要结构，包含会话的所有操作
type Session struct {
db *sql.DB // 数据库连接
sql strings.Builder // sql 用于拼接 SQL 语句
sqlVars []interface{} // sqlVars 用于存储 SQL 语句中的参数
dialect dialect.Dialect // dialect 记录了该 Session 所使用的数据库方言
refTable *schema.Schema // refTable 记录 Model 对应的表结构
clause clause.Clause // clause 是记录 SQL 语句中的各种子句
tx *sql.Tx // tx 提供事务支持，如果 tx 不为 nil，则执行所有操作都在事务中
}
```

### 5.1 CRUD 实现

`Session` 提供了以下核心方法：

- **Insert**: 插入记录到数据库。
- **Find**: 查找记录并填充到指定结构体中。
- **Update**: 更新记录。
- **Delete**: 删除记录。
- **Count**: 返回记录总数。
- **First**: 返回第一条记录。

这些方法通过组合 SQL 子句（如 `INSERT`, `SELECT`, `WHERE` 等）生成 SQL 语句，并执行数据库操作。

`Session` 使用 `clause` 字段管理 SQL 子句（如 `INSERT`, `SELECT`, `WHERE` 等）。通过 `Set` 方法设置子句，通过 `Build` 方法生成完整的 SQL 语句。

`Session` 使用 `refTable` 字段管理表结构（`Schema`）。通过 `Model` 方法设置当前操作的模型，通过 `RefTable` 方法获取表结构。

`Session` 提供了以下方法执行 SQL 语句：

- **Exec**: 执行不需要返回行的 SQL 语句（如 `INSERT`、`UPDATE`）。
- **QueryRow**: 执行返回单行结果的 SQL 查询。
- **QueryRows**: 执行返回多行结果的 SQL 查询。

这些方法通过 `Raw` 方法构建 SQL 语句，并调用底层的 `database/sql` 接口执行。

> [!NOTE]
> Exec 方法用于执行不需要返回行的 SQL 语句，例如 INSERT、DELETE、UPDATE 等；返回值 result 是 sql.Result 类型，用于返回执行结果（受影响的行数等）。
> QueryRow 方法用于执行需要返回单行结果的 SQL 查询，例如 SELECT 查询，返回一个 \*sql.Row 对象，表示查询结果的单行记录。

### 5.2 事务的实现

数据库事务(transaction)是访问并可能操作各种数据项的一个数据库操作序列，这些操作要么全部执行,要么全部不执行，是一个不可分割的工作单位。事务由事务开始与事务结束之间执行的全部数据库操作组成。

> [!NOTE] ACID
> 原子性(Atomicity)：事务中的全部操作在数据库中是不可分割的，要么全部完成，要么全部不执行。
> 一致性(Consistency): 几个并行执行的事务，其执行结果必须与按某一顺序 串行执行的结果相一致。
> 隔离性(Isolation)：事务的执行不受其他事务的干扰，事务执行的中间结果对其他事务必须是透明的。
> 持久性(Durability)：对于任意已提交事务，系统必须保证该事务对数据库的改变不被丢失，即使数据库出现故障。

在 SQLite 中，`BEGIN` 开启事务，`COMMIT` 提交事务，`ROLLBACK` 回滚事务。任何一个事务，均以 `BEGIN` 开始，`COMMIT` 或 `ROLLBACK` 结束。Go 标准库 database/sql 也提供了支持事务的接口。

调用 `db.Begin()` 得到 `*sql.Tx` 对象，使用 `tx.Exec()` 执行一系列操作，如果发生错误，通过 `tx.Rollback()` 回滚，如果没有发生错误，则通过 `tx.Commit()` 提交。

我们发现，一般是用 db 来调用的，但是在事务中是是用 tx 来调用，因此，我们将他封装一下，当 `tx` 不为空时，则使用 `tx` 执行 SQL 语句，否则使用 `db` 执行 SQL 语句。这样既兼容了原有的执行方式，又提供了对事务的支持。

```go
// 抽象出一个接口 CommonDB，包含 Query、QueryRow、Exec 三个方法
type CommonDB interface {
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    Exec(query string, args ...interface{}) (sql.Result, error)
}

// 该接口的实现有 *sql.DB 和 *sql.Tx
var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// DB 返回 *sql.DB 对象
func (s *Session) DB() CommonDB {
    if s.tx != nil {
        return s.tx
    }
    return s.db
}
```

这样，对于 Exec 方法，我们既不用 `db.Exec()` 也不用 `tx.Exec()`，而是使用 `DB().Exec()`。

接着，我们封装一个 Transaction 方法

```go
// 使用一个这种接口函数封装事务内的操作
type TxFunc func(*session.Session) (interface{}, error)  

// 用于执行一个事务，无论事务是否失败，可以通过 defer 中的回滚确保事务的正确结束
// s.Begin()、s.Rollback()、s.Commit() 简单封装了一下标准库函数
func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {  
 s := engine.NewSession()  
 if err := s.Begin(); err != nil {  
  return nil, err  
 }  
 defer func() {  
  if p := recover(); p != nil {  
   _ = s.Rollback()  
   panic(p) // 在回滚之后继续抛出 panic
  } else if err != nil {  
   _ = s.Rollback() // 回滚
  } else {  
   err = s.Commit() // 提交
  }  
 }()  
 return f(s)  
}
```

### 5.3 Hooks 钩子函数

钩子机制允许在数据库操作（如查询、插入、更新、删除）前后执行自定义逻辑。

`CallMethod` 方法用于调用钩子函数。其核心逻辑如下：

- 根据方法名（如 `BeforeQuery`）和可选的值（如模型实例）查找钩子函数。
- 如果找到钩子函数，则调用它，并传递当前的 `Session` 作为参数。

钩子机制为 ORM 提供了扩展性，用户可以在数据库操作前后插入自定义逻辑，例如日志记录、数据校验等。例如，我们可以在查询语句后，调用钩子函数，将密码修改为 `****` 这种形式，从而避免用户直接读取。

## 6 数据库迁移

GeeORM 的 `Migrate` 操作仅针对最为简单的场景，即支持字段的新增与删除，不支持字段类型变更。

首先，判断是否已经存在该表。如果不存在，则创建表。如果存在，则需要新增表中缺少的字段，并删除表中多余的字段。新增或删除字段的过程，可以使用事务来保证原子性。
