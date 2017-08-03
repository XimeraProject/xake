# Installing

If you haven't installed [ximeraLatex](https://github.com/XimeraProject/ximeraLatex) you should do so.

## Install on Archlinux

The AUR contains [xake-git](https://aur.archlinux.org/packages/xake-git/) which will install the latest version of xake.  Therefore
```
yaourt -S xake-git
```
should suffice if you use yaourt.

## Install on Ubuntu and Debian

Download the `.deb` from the [releases link above](https://github.com/XimeraProject/xake/releases).  Then install it with
```
sudo dpkg -i xake_0.8.20_amd64.deb
```

## Install on Redhat and CentOS

Download the `.rpm` from the [releases link above](https://github.com/XimeraProject/xake/releases).  Then install it with
```
sudo rpm -i xake-0.8.20.x86_64.rpm
```

## Install from source

If you would rather compile xake yourself, the following should work.
```
mkdir -p ~/go/src/github.com/ximeraproject
export GOPATH=$HOME/go
cd ~/go/src/github.com/ximeraproject
git clone https://github.com/XimeraProject/xake.git
cd xake
go get .
go build .

```
