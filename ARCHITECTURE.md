Architecture
===

Forwarder
---

The forwarding mechanism orchestrates Kafka consumption, an (optional) HTTP
server, message file uploads, and (optional) uploads of report files through a
set of cooperating goroutines coordinated by contexts and buffered channels.

### Normal Execution

```mermaid
flowchart TD
    start([start])
    start --> Exec
    subgraph Exec["Executor"]
        setup["Setup contexts, signal handling, channels"]
        setup --> startServer["Start HTTP server (if requested)"]
        startServer --> startForwarder["Start forwarder goroutine"]
        startForwarder --> waitForExitSignal["Wait for exit signal"]
    end

    server["HTTP server"]
    startServer --> server
    startForwarder --> setupF1
    

    workChan[[Upload work channel]]
    resultChan[[Upload result channel]]

    subgraph Forwarder["Forwarder"]
        setupF1["Start upload workers"]
        setupF1 --> setupF2["Start reporter worker (if requested)"]
        setupF2 --> setupQueue["Set up work queue"]
        setupQueue --> checkQueue
            
        subgraph mainLoop["Main Loop"]
            checkQueue{Is work queue empty?}
            checkQueue --> |yes| fetch["Fetch Kafka records"]
            fetch --> putInQueue["Put work in queue"]
            checkQueue --> |no| sendToWorker["Send to worker"]
            sendToWorker --> keepUnpushed["Keep unpushed work for later"]
            putInQueue --> switchProfile["Switch to next profile (if applicable)"]
            keepUnpushed --> switchProfile["Switch to next profile (if applicable)"]
            switchProfile --> checkQueue
        end
    end

    sendToWorker --> |Try sending work| workChan

    subgraph uploadWorker1["Upload Worker (1)"]
        dest1["Upload to destination (with retries)"]
    end

    subgraph uploadWorker2["Upload Worker (2)"]
        dest2["Upload to destination (with retries)"]
    end
    
    subgraph uploadWorkerN["Upload Worker (N)"]
        destN["Upload to destination (with retries)"]
    end

    workChan --> uploadWorker1
    workChan --> uploadWorker2
    workChan --> uploadWorkerN

    uploadWorker1 -->|Send result| resultChan
    uploadWorker2 -->|Send result| resultChan
    uploadWorkerN -->|Send result| resultChan

    resultChan --> reporterWorker
    subgraph reporterWorker["Reporter Worker"]
        store["Store result per topic in an in-memory buffer"]
        store --> reportUploads["Upload report to destination once a buffer reaches configured limit"]
        reportUploads --> store
    end
```

### Shutdown Mechanism

```mermaid
sequenceDiagram
    participant OS
    participant Exec as Executor
    participant Server as HTTP server
    participant Forwarder
    participant Workers as Upload workers
    participant Reporter as Reporter worker

    OS->>Exec: SIGINT
    Exec->>Forwarder: Send shutdown signal (via context)
    Forwarder->>Workers: Send pending work (if applicable)
    Forwarder->>Workers: Once all work is sent, send shutdown signal (via context)
    Workers->>Workers: Pull in any pending work, finish it, then exit
    Forwarder->>Reporter: Send shutdown signal (via context)
    Reporter->>Reporter: Pull in any pending results and handle them
    Reporter->>Reporter: Upload pending report files concurrently
    Reporter->>Reporter: Wait for report uploads to finish
    Reporter->>Reporter: Signal exit to forwarder (via channel)
    Forwarder->>Forwarder: Exit
    Exec->>Server: Send shutdown signal (via context)
    Server->>Server: Exit
    Exec->>Exec: Exit
```

Note: The executor has a timeout for graceful shutdown. If the running
components do not shut down till that timeout is reached, it exits forcefully.
If a second shutdown signal is sent by the OS, the executor immediately exits as
well.
