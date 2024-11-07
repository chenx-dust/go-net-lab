package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type WaitingPing struct {
	startTime time.Time
	realId    int
	data      *[]byte
}

func main() {
	// Create channel for signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create done channel for coordination
	done := make(chan bool)

	// Parse flags
	dst := flag.String("dst", "127.0.0.1:9000", "destination address")
	tcp := flag.Bool("tcp", false, "use tcp")
	udp := flag.Bool("udp", false, "use udp")
	count := flag.Int("count", 0, "number of pings")
	interval := flag.Duration("interval", 1*time.Second, "interval between pings")
	pkgSize := flag.Int("size", 1024, "packet size")
	flag.Parse()

	if *pkgSize < 3 {
		log.Fatalf("Packet size must be greater than 2")
	}

	pingRecord := make([]time.Duration, 0, *count)
	waitingPing := make(map[int]WaitingPing)
	mutex := sync.Mutex{}

	start := time.Now()

	// Start ping operations in goroutine
	go func() {
		if *tcp {
			pingTcp(*dst, *count, *interval, *pkgSize, &pingRecord, &waitingPing, &mutex)
		} else if *udp {
			pingUdp(*dst, *count, *interval, *pkgSize, &pingRecord, &waitingPing, &mutex)
		} else {
			log.Fatalf("No protocol specified")
		}
		done <- true
	}()

	// Wait for either completion or interrupt
	select {
	case <-done:
		// Normal completion
	case sig := <-sigChan:
		log.Printf("Received signal: %v, shutting down...", sig)
	}

	avg := time.Duration(0)
	min := pingRecord[0]
	max := pingRecord[0]
	sqrAvg := time.Duration(0)
	loss := 0
	for _, d := range pingRecord {
		if d <= 0 {
			loss++
			continue
		}
		avg += d
		sqrAvg += d * d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}
	num := len(pingRecord) - loss
	if num == 0 {
		log.Fatalf("No pings received")
	}
	avg /= time.Duration(num)
	stddev := time.Duration(math.Sqrt(float64(sqrAvg/time.Duration(num) - avg*avg)))

	fmt.Println("================================================")
	fmt.Printf("Time elapsed: %s\n", time.Since(start))
	fmt.Printf("Transmitted: %d, Received: %d, Loss: %f%%\n", len(pingRecord), num, 100*float64(loss)/float64(len(pingRecord)))
	fmt.Printf("RTT Min: %s, Avg: %s, Max: %s, Stddev: %s\n", min, avg, max, stddev)
}

func consumePing(waitingPings *map[int]WaitingPing, pingRecord *[]time.Duration, buffer *[]byte, mutex *sync.Mutex) error {
	if (*buffer)[0] != 0x8a {
		return fmt.Errorf("invalid magic number")
	}
	id := int((*buffer)[1]) + int((*buffer)[2])*256
	mutex.Lock()
	defer mutex.Unlock()
	ping, exists := (*waitingPings)[id]
	if !exists {
		return fmt.Errorf("ping with id %d does not exist", id)
	}
	duration := time.Since(ping.startTime)
	if !bytes.Equal(*ping.data, *buffer) {
		return fmt.Errorf("data mismatch")
	}
	log.Printf("Ping #%d, RTT: %s", ping.realId, duration)
	(*pingRecord)[ping.realId] = duration
	delete(*waitingPings, id)
	return nil
}

func ping(conn net.Conn, count int, interval time.Duration, pkgSize int, pingRecord *[]time.Duration, waitingPings *map[int]WaitingPing, mutex *sync.Mutex) {
	go func() {
		buffer := make([]byte, pkgSize)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				log.Printf("Failed to read: %v", err)
				continue
			}
			if n != pkgSize {
				log.Printf("Received %d bytes, expected %d", n, pkgSize)
				continue
			}
			err = consumePing(waitingPings, pingRecord, &buffer, mutex)
			if err != nil {
				log.Printf("Failed to consume ping: %v", err)
			}
		}
	}()

	buffer := make([]byte, pkgSize)
	for i := 0; i < count || count == 0; i++ {
		// randomize buffer content
		rand.Read(buffer)
		buffer[0] = 0x8a
		buffer[1] = byte(i)
		buffer[2] = byte(i / 256)
		mutex.Lock()
		*pingRecord = append(*pingRecord, 0)
		(*waitingPings)[i] = WaitingPing{startTime: time.Now(), realId: i, data: &buffer}
		mutex.Unlock()

		go func() {
			_, err := conn.Write(buffer)
			if err != nil {
				log.Printf("Failed to write: %v", err)
			}
		}()
		time.Sleep(interval)
	}
}

func pingTcp(dst string, count int, interval time.Duration, size int, pingRecord *[]time.Duration, waitingPings *map[int]WaitingPing, mutex *sync.Mutex) {
	fmt.Println("Pinging", dst, "with TCP")
	fmt.Println("Count:", count)
	fmt.Println("Interval:", interval)
	fmt.Println("Packet size:", size)
	fmt.Println("================================================")

	conn, err := net.Dial("tcp", dst)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	ping(conn, count, interval, size, pingRecord, waitingPings, mutex)
}

func pingUdp(dst string, count int, interval time.Duration, size int, pingRecord *[]time.Duration, waitingPings *map[int]WaitingPing, mutex *sync.Mutex) {
	fmt.Println("Pinging", dst, "with UDP")
	fmt.Println("Count:", count)
	fmt.Println("Interval:", interval)
	fmt.Println("Packet size:", size)
	fmt.Println("================================================")

	udpAddr, err := net.ResolveUDPAddr("udp", dst)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	ping(conn, count, interval, size, pingRecord, waitingPings, mutex)
}
