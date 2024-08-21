package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func openRawIps(filePath string, port int) ([]string, error) {
	input, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer input.Close()
	var formatedIPS []string
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		ipRanges := strings.Split(line, ",")
		for _, v := range ipRanges {
			ipRange := strings.TrimSpace(v)
			parts := strings.Split(ipRange, "-")
			if len(parts) == 2 {
				startIP := fmt.Sprintf("%s:%d", strings.TrimSpace(parts[0]), port)
				endIP := fmt.Sprintf("%s:%d", strings.TrimSpace(parts[1]), port)
				formatedIPS = append(formatedIPS, fmt.Sprintf("%s-%s", startIP, endIP))
			} else {
				return nil, fmt.Errorf("invalid ip range: %s", ipRange)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return formatedIPS, nil
}

func isReachable(address string) bool {
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
}

func scanIpRange(ipRanges []string, wg *sync.WaitGroup) error {
	result, err := os.OpenFile("result.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer result.Close()

	for _, ipRange := range ipRanges {
		parts := strings.Split(ipRange, "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid ip range: %s", ipRange)
		}
		startIP := parts[0]
		endIP := parts[1]
		start := net.ParseIP(strings.Split(startIP, ":")[0])
		end := net.ParseIP(strings.Split(endIP, ":")[0])
		if start == nil || end == nil {
			return fmt.Errorf("invalid ip range: %s", ipRange)
		}
		for ip := start; !ip.Equal(end); incrementIP(ip) {
			wg.Add(1)
			address := fmt.Sprintf("%s:%s", ip.String(), strings.Split(startIP, ":")[1])
			fmt.Println(address)
			go func(addr string) {
				defer wg.Done()
				if isReachable(addr) {
					if _, err := result.WriteString(fmt.Sprintf("%s\n", addr)); err != nil {
						panic(err)
					}
				}
			}(address)
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		panic("Please provide an input file")
	}
	var wg sync.WaitGroup
	ips, err := openRawIps(os.Args[1], 3389)
	if err != nil {
		panic(err)
	}
	if err := scanIpRange(ips, &wg); err != nil {
		panic(err)
	}
	wg.Wait()
}
