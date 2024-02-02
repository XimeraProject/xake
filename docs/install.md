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

If you would rather compile xake yourself, the following may work.
```
mkdir -p ~/go/src/github.com/ximeraproject
export GOPATH=$HOME/go
cd ~/go/src/github.com/ximeraproject
git clone https://github.com/XimeraProject/xake.git
cd xake
go install .
```

That may not work, though, depending on your version of libgit2.  To build libgit2 statically, you could instead follow the following recipe:
```
export GOPATH=$HOME/go
export PKG_CONFIG_PATH=$HOME/go/src/github.com/libgit2/git2go/vendor/libgit2/build
export CGO_CFLAGS=-I$HOME/go/src/github.com/libgit2/git2go/vendor/libgit2/include
mkdir -p ~/go/src/github.com/ximeraproject
cd ~/go/src/github.com/ximeraproject
git clone https://github.com/XimeraProject/xake.git
go get -d github.com/libgit2/git2go
cd ~/go/src/github.com/libgit2/git2go
git submodule update --init
make install-static
cd ~/go/src/github.com/ximeraproject/xake
go get -tags static .
```

Then in the directory `~/go/bin` you should find a `xake` binary.
