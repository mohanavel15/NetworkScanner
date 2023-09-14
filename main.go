package main

import (
	"fmt"
	"net"
	"sync"
)

const (
	Red   = "\033[31m"
	Green = "\033[32m"
	Blue  = "\033[34m"
	Reset = "\033[0m"
)

func main() {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var filteredInterfaces []net.Interface
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 {
			filteredInterfaces = append(filteredInterfaces, iface)
		}
	}

	fmt.Println(Blue + "Select a network interface:" + Reset)
	for i, iface := range filteredInterfaces {
		fmt.Printf("%s%d. %s%s\n", Blue, i+1, iface.Name, Reset)
	}

	var choice int
	fmt.Print(Green + "Enter the number of the interface: " + Reset)
	_, err = fmt.Scanf("%d", &choice)
	if err != nil || choice < 1 || choice > len(filteredInterfaces) {
		fmt.Println(Red + "Invalid choice" + Reset)
		return
	}

	selectedInterface := filteredInterfaces[choice-1]

	addrs, err := selectedInterface.Addrs()
	if err != nil {
		fmt.Println(Red+"Error:", err.Error(), Reset)
		return
	}

	fmt.Printf("%sIP addresses for interface %s:%s\n", Blue, selectedInterface.Name, Reset)

	v4Addrs := []net.Addr{}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			fmt.Printf("%s%s%s\n", Green, addr.String(), Reset)
			v4Addrs = append(v4Addrs, addr)
		}
	}

	if len(v4Addrs) <= 0 {
		fmt.Println(Red + "No IPv4 Address Found!" + Reset)
		return
	}

	fmt.Println(Blue + "Hosts found in the given IP range:" + Reset)

	ipRange := v4Addrs[0].String()

	ips, err := getIPsInRange(ipRange)
	if err != nil {
		fmt.Println(Red+"Error:", err, Reset)
		return
	}

	var wg sync.WaitGroup

	for _, ip := range ips {
		wg.Add(1)
		go ScanIP(&wg, ip)
	}

	wg.Wait()
}

func ScanIP(wg *sync.WaitGroup, ip string) {
	defer wg.Done()
	for i := 0; i < 3; i++ {
		hostname, err := resolveHostname(ip)
		if err != nil {
			continue
		}
		fmt.Printf("%sIP %s: Hostname - %s%s\n", Green, ip, hostname, Reset)
		break
	}
}

func getIPsInRange(ipRange string) ([]string, error) {
	ip, ipNet, err := net.ParseCIDR(ipRange)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	ips = ips[1 : len(ips)-1]

	return ips, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func resolveHostname(ip string) (string, error) {
	names, err := net.LookupAddr(ip)
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "N/A", nil
	}
	return names[0], nil
}
