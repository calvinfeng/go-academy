# Go Concurrency Part 1

Now we know what a goroutine is. Here are some fairly advanced use cases. The majority of the examples come from Rob Pike's talk. I made a couple of enhancements here and there. Try your best to understand them, but know that concurrency is difficult and it takes time to have a firm grasp.

## Channels

Two goroutines can communicate and synchronize through channel.

```go
func ping(msg string, c chan string) {
  for i := 0; ; i++ {
    c <- fmt.Sprintf("%s %d", msg, i)
    time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
  }
}

func main() {
  ch := make(chan string)
  go ping("ping", ch)

  for i := 0; i < 5; i++ {
    fmt.Println(<-ch)
    // The receiver end is blocking because it's waiting for a message to come through the channel.
  }

  fmt.Println("Enough pings, done!")
}
```

## Generator

What if we want to listen to multiple pings? We can use a ping generator.

```go
func pingGen(msg string) <-chan string {
  ch := make(chan string)

  go func() {
    for i := 0; ; i++ {
      ch <- fmt.Sprintf("%s %d", msg, i)
      time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
    }
  }()

  return ch
}

func main() {
  ch1 := pingGen("ping")
  ch2 := pingGen("bing")

  for i := 0; i < 5; i++ {
    fmt.Println(<-ch1)
    fmt.Println(<-ch2)
  }

  fmt.Println("Enough pings, done!")
}
```

Here's a problem though, `ch1` and `ch2` are blocking each other. We are not really getting the live updates of ping.

## Multiplexing

We can address the problem by multiplexing multiple channels into one.

```go
func fanIn(inputs ...<-chan string) <-chan string {
  out := make(chan string)

  for _, in := range inputs {
    go func(ch <-chan string) {
      for {
        out <- <-ch
      }
    }(in)
  }

  return out
}

func pingGen(msg string) <-chan string {
  ch := make(chan string)

  go func() {
    for i := 0; ; i++ {
      ch <- fmt.Sprintf("%s %d", msg, i)
      time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
    }
  }()

  return ch
}

func main() {
  out := fanIn(pingGen("ping"), pingGen("bing"), pingGen("sing"))

  for i := 0; i < 10; i++ {
    fmt.Println(<-out)
  }

  fmt.Println("Enough pings, done!")
}
```

However, we can easily spot that the messages are coming in out of order. What if I want the batch of messages to come in order, e.g.

```text
sing 0
bing 0
ping 0
sing 1
ping 1
bing 1
sing 2
ping 2
bing 2
sing 3
ping 3
bing 3
```

## Fan-in Sequencing

We will use a signal channel to achieve the wait. We wait for three pings and then tell them ready for the next batch of 3 pings. First, let's define a message struct.

```go
type message struct {
  content string
  ready   chan bool
}
```

Each message will carry a reference to a signal channel, called `ready`. When the channel receives a signal, it indicates that it's ready to process next message.

```go
func fanIn(inputs ...<-chan message) <-chan message {
  out := make(chan message)

  for _, in := range inputs {
    go func(ch <-chan message) {
      for {
        out <- <-ch
      }
    }(in)
  }

  return out
}

func pingGen(msg string) <-chan message {
  ch := make(chan message)
  rdy := make(chan bool)

  go func(ready chan bool) {
    for i := 0; ; i++ {
      ch <- message{fmt.Sprintf("%s %d", msg, i), ready}
      time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
      <-ready // Wait for ready
    }
  }(rdy)

  return ch
}

func main() {
  out := fanIn(pingGen("ping"), pingGen("bing"), pingGen("sing"))

  for i := 0; i < 10; i++ {
    msgs := []message{}
    for j := 0; j < 3; j++ {
      msgs = append(msgs, <-out)
    }

    // Grab three messages and then tell each generator that it's ready for next message.
    for _, msg := range msgs {
      fmt.Println(msg.content)
      msg.ready <- true
    }
  }

  fmt.Println("Enough pings, done!")
}
```

## Select

If you already know how many channels you want to listen to, it's better to use `select` instead of fan-in. However, it's also reasonable to combine two techniques together. Let's say I want a timeout on my fan-in.

```go
func main() {
  out := fanIn(pingGen("ping"), pingGen("ding"), pingGen("sing"))
  timeout := time.After(5 * time.Second)

  for {
    select {
      case msg := <-out:
        fmt.Println(msg.content)
        msg.ready <- true
      case <-timeout:
        fmt.Println("Too late")
        return
    }
  }
}
```

## Quit Channel

We know how to terminate the reader, but how about the writer? Notice that despite the for-loop has terminated, the goroutine in each generator is not terminated. Let's do something clever, i.e. use a channel to signal quit. We need to stop using `fanIn` for a moment because that will cause a deadlock if we terminate the writer without terminating the goroutines from `fanIn`.

```go
func pingGen(msg string, quit chan bool) <-chan message {
  ch := make(chan message)
  rdy := make(chan bool)

  go func(ready chan bool) {
    for i := 0; ; i++ {
      select {
      case ch <- message{fmt.Sprintf("%s %d", msg, i), ready}:
        time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
        <-ready
      case <-quit:
        fmt.Printf("stopped sending %s\n", msg)
        return
      }
    }
  }(rdy)

  return ch
}
```

Now we can tell it to stop.

```go
func main() {
  stopPing := make(chan bool)

  ping := pingGen("ping", stopPing)

  for i := 0; i < 10; i++ {
    msg := <-ping
    fmt.Println(msg.content)
    msg.ready <- true
  }

  fmt.Println("Enough pings, stop it!")

  stopPing <- true
}
```

## Receive on Quit

How do we know the quit signal has been received and processed? We can wait for the quit channel to tell us that it is done! We make a channel that passes a channel which passes a string.

```go
func pingGen(msg string, quit chan chan string) <-chan message {
  ch := make(chan message)
  rdy := make(chan bool)

  go func(ready chan bool) {
    for i := 0; ; i++ {
      select {
      case ch <- message{fmt.Sprintf("%s %d", msg, i), ready}:
        time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
        <-ready
      case reply := <-quit:
        reply <- fmt.Sprintf("stopped sending %s\n", msg)
        return
      }
    }
  }(rdy)

  return ch
}
```

We would create a quit channel that passes a reply channel.

```go
func main() {
  stopPing := make(chan chan string)

  ping := pingGen("ping", stopPing)

  for i := 0; i < 10; i++ {
    msg := <-ping
    fmt.Println(msg.content)
    msg.ready <- true
  }

  fmt.Println("Enough pings, stop it!")

  reply := make(chan string)
  stopPing <- reply
  fmt.Println(<-reply)
}
```

## Daisy Chain

Let's try to pass data from one channel to another, and repeat this process many times. We are daisy chaining go routines.

For each pass, we increment the value of an integer by one.

```go
func pass(left, right chan int) {
  value := <-left
  right <- value + 1
  fmt.Printf("Left[%d] -> Right[%d]\n", value, value + 1)
}
```

Now create the goroutines and chain them up.

```go
func main() {  
  end := make(chan int)

  var left chan int
  var right chan int

  right = end
  for i := 0; i < 6; i++ {
    left = make(chan int)
    go pass(left, right)
    right = left
  }

  fmt.Println("All goroutines are waiting.")

  // Send the initial value  to first channel.
  initVal := 1
  go func(ch chan int, val int) {
    fmt.Println("Give left most channel the initial value")
    ch <- val
  }(left, initVal)

  fmt.Printf("Final value is %d\n", <-end)
}
```

