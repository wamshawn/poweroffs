# poweroffs
one application use unix or tcp listener to listen power off signal to power off linux system.


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

## Build service
Clone project.
```shell
git clone https://github.com/wamshawn/poweroffs.git

cd poweroffs
```
### DEB

Build bin.
```shell
go build -o ./assemble/deb/poweroffs/usr/local/bin/poweroffs
```

Into deb.
```shell
cd ./assemble/deb
```

Chmod +X.
```shell
chmod +x ./poweroffs/DEBIAN/postinst
chmod +x ./poweroffs/DEBIAN/prerm
```

Package
```shell
sudo apt install fakeroot
fakeroot dpkg-deb --build poweroffs
```

Install
```shell
sudo dpkg -i poweroffs
```

Uninstall
```shell
dpkg -r poweroffs
```