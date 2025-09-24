## Go协程

## 1. 什么是协程（goroutine）

- **goroutine** 是 Go 语言中由 **Go 运行时调度器（runtime scheduler）** 管理的轻量级线程。
- 比操作系统线程（thread）更轻量，启动成本极低。
- 一个 Go 程序里可以同时运行 **成千上万个 goroutine**。

👉 启动协程只需要在函数调用前加上 `go` 关键字。

------

## 2. 基本用法

```go
package main

import (
	"fmt"
	"time"
)

func hello() {
	fmt.Println("Hello from goroutine")
}

func main() {
	go hello() // 启动一个协程执行 hello()
    
	// 故意加个耗时循环（可以调整循环次数观看效果）
	for i := 0; i < 1e6; i++ {
	}
	fmt.Println("Hello from main")

	// 等一会，不然主协程退出，子协程可能来不及执行
	time.Sleep(time.Second)
}
```

输出可能是：

```
Hello from main
Hello from goroutine
```

也可能是：

```
Hello from goroutine
Hello from main
```

这是因为协程的执行顺序不确定，调度由 Go runtime 决定

------

## 3. 协程 vs 线程

| 特性     | 协程（goroutine）           | 线程（thread）    |
| -------- | --------------------------- | ----------------- |
| 启动开销 | 小（几 KB 栈空间）          | 大（MB 级别内存） |
| 数量     | 支持成千上万个              | 通常几百到几千    |
| 调度     | Go runtime 调度（M:N 模型） | 操作系统调度      |
| 切换速度 | 快（用户态）                | 慢（内核态）      |

------

## 4. 协程通信 —— channel

Go 提倡 **不要通过共享内存来通信，而是通过通信来共享内存**。
 这就要用到 **channel**。

```go
package main

import (
	"fmt"
)

func worker(ch chan string) {
	ch <- "Hello from worker" // 往通道里发消息
}

func main() {
	ch := make(chan string)

	go worker(ch)

	msg := <-ch // 从通道里取消息
	fmt.Println(msg)
}
```

输出：

```
Hello from worker
```

------

## 5. 实际应用场景

- 并发网络请求
- 消息队列消费者（配合 Kafka、RabbitMQ 等）
- 定时任务、后台任务
- 高并发 Web 服务器

------

👉 总结一句：
 Go 的协程（goroutine）= **廉价的并发执行单元**，搭配 `channel` 实现高效的并发编程，是 Go 语言最大的优势之一。