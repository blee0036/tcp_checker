# README.md

## TCP Checker

TCP Checker is a simple tool that allows you to check the TCP connection status of a specified host and port. It attempts to connect to the target host multiple times and records the number of successful connections, the average connection time, and the ratio of lost connections.

## Installation

You can download the latest binary file from the [release page](https://github.com/blee0036/tcp_checker/releases/). After downloading, unzip the file to get the executable file `tcp_checker`.

For example, for Linux amd64 systems, you can download `tcp_checker_vX.X.X_linux_amd64.tar.gz`, then unzip it:

```bash
tar -xzvf tcp_checker_vX.X.X_linux_amd64.tar.gz
```

## Usage

Run `tcp_checker`, it listens on port 8080 by default, makes 5 connection attempts with no token require:

```bash
./tcp_checker
```

You can check target tcp connection with host and port at the query parameters `host`, `port` and `token`:

```bash
curl "http://localhost:8080/?host=example.com&port=80&token=<your-token>"
```

You can also batch check tcp connection by post the data :

```bash
curl --location 'http://localhost:8080/batch?token=<your-token>' \
--header 'Content-Type: text/plain' \
--data '1.1.1.1:80
google.com:443
10.0.0.1:1234'
```

### Parameter Explanation:

- `-p, --port <port>`: Port to listen on (default 8080)
- `-a, --attempts <attempts>`: Number of connection attempts to a TCP port (default 5)
- `-tO, --timeoutMS <timeoutMS>`: Timeout in milliseconds for each connection detection (default 2000)
- `-t, --token <token>`: Authentication token for requests (default empty)

### Simple Example
```bash
./tcp_checker -p 8081 -a 10 -tO 3000 -t "Ver1f^"
```

## Run as a Service

On a Linux system, you can create a systemd service to manage `tcp_checker`. First, create a new systemd service file:

```bash
sudo nano /etc/systemd/system/tcp_checker.service
```

In the editor that opens, enter the following content:

```ini
[Unit]
Description=TCP Checker
After=network.target

[Service]
ExecStart=/path/to/tcp_checker
Restart=on-abnormal
RestartSec=5s
StandardOutput=null
StandardError=syslog

[Install]
WantedBy=multi-user.target
```

In the `ExecStart` line, `/path/to/tcp_checker` should be replaced with the actual path of the `tcp_checker` executable file.

Save and exit the editor. Then, reload daemon for service:
```bash
systemctl daemon-reload
```

Start the `tcp_checker` service

```bash
systemctl start tcp_checker
```

If everything is fine, you can set the `tcp_checker` service to start on boot:

```bash
systemctl enable tcp_checker
```

Now, you can manage the `tcp_checker` service with the `systemctl` command. For example, to check the service status:

```bash
systemctl status tcp_checker
```