#!/usr/bin/env python3
import socket
import sys
import concurrent.futures

# Common ports to check (simplified)
COMMON_PORTS = {
    21: "FTP",
    22: "SSH",
    23: "Telnet",
    25: "SMTP",
    53: "DNS",
    80: "HTTP",
    110: "POP3",
    135: "RPC",
    139: "NetBIOS",
    143: "IMAP",
    443: "HTTPS",
    445: "SMB",
    3306: "MySQL",
    3389: "RDP",
    5432: "PostgreSQL",
    6379: "Redis",
    8080: "HTTP-Alt"
}

def scan_port(ip, port, timeout=1.0):
    try:
        with socket.socket(socket.AF_INET, socket.socket.SOCK_STREAM) as s:
            s.settimeout(timeout)
            result = s.connect_ex((ip, port))
            if result == 0:
                return port, True
            else:
                return port, False
    except Exception:
        return port, False

def scan_target(target_ip):
    print(f"Scanning target {target_ip} for open common ports...")
    open_ports = []
    
    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
        futures = {executor.submit(scan_port, target_ip, port): port for port in COMMON_PORTS.keys()}
        
        for future in concurrent.futures.as_completed(futures):
            port, is_open = future.result()
            if is_open:
                service = COMMON_PORTS[port]
                print(f"[!] Port {port} is OPEN ({service})")
                open_ports.append(port)
                
    if not open_ports:
        print("No common ports found open.")
    else:
        print(f"\nWarning: {len(open_ports)} open ports found. Ensure these are intended to be public.")
        
    # Security check: DB or internal services exposed?
    dangerous_ports = {3306, 5432, 6379, 3389, 23, 21}
    exposed = set(open_ports).intersection(dangerous_ports)
    if exposed:
        print("\nCRITICAL: Dangerous services (Database/RDP/Telnet) are publicly exposed!")
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: port_scanner.py <IP_ADDRESS>")
        sys.exit(1)
        
    target = sys.argv[1]
    scan_target(target)
