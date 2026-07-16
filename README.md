# poweroffs
one application use unix or tcp listner to listen power off signal to power off linux system.


## Install

```shell
go install github.com/wamshawn/poweroffs@latest
```

## Usage

### Listen

Use unix sock only.
```shell
sudo run poweroffs --unix=poweroffs.sock
```

With tcp, ca is required.
```shell
sudo run poweroffs --unix=poweroffs.sock \
 --tcp=127.0.0.1:13000 \
 --cert=ca.crt \
 --key=ca.key
```

### Send signal

Connect to server and send `reboot` or `poweroff`, the target system will restart or shutdown after 500ms.

