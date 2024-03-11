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

You can specify the target host and port with the query parameters `host`, `port` and `token`:

```bash
curl "http://localhost:8080/?host=example.com&port=80&token=<your-token>"
```

You can also specify the listening port and the number of connection attempts with the `-p` and `-a` options:
```bash
./tcp_checker -p 8081 -a 10 -t VeR1fy
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