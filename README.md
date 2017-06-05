# Xake

[![Tag](https://img.shields.io/github/tag/XimeraProject/xake.svg?style=flat-square)](https://github.com/XimeraProject/xake/tags)
[![License: GPL v3](https://img.shields.io/badge/license-GPL%20v3-blue.svg?style=flat-square)](https://github.com/XimeraProject/xake/blob/master/LICENSE.md)

Xake is the build tool for Ximera.  It is used to manage the
conversion of TeX files into .html files, and the publication of the
resulting .html files.

The basic workflow is as follows.

1) Make some edits.
2) `git add` and `git commit` to commit your source TeX files to your repository.
3) `xake bake` compiles your TeX files into .html files.
4) `xake frost` creates a special git tag for the .html files.
5) `xake serve` pushes your TeX source and the special git tag to the Ximera server.

The `xake bake` step is smart enough to only recompile files which
have changed.  The `xake frost` step creates the "frosting" meaning a
git tag pointing to a commit sitting on top of the repo's HEAD.  The
final `xake serve` is actually just a wrapper around `git push` which
pushes the frosting to the server.

## Getting Started

### Installing

If you haven't installed `ximeraLatex` you should do so.

### Using xake

First, if you don't already have a GPG key, create one.  You can get
started with `gpg --gen-key` and then following instructions such as
<https://help.github.com/articles/generating-a-new-gpg-key/>.

To be permitted to make use the `ximera.osu.edu` server, you will need
to have your GPG key trusted by the Ohio State team.  Email
ximera@math.osu.edu with a copy of your key and we'll sign it for you;
alternatively, find someone else who has permission to use
`ximera.osu.edu` and simply have them sign your key.

To share your signed public key with Ximera, use the command

`gpg --keyserver hkps://ximera.osu.edu/ --send-key YOUR-GPG-KEY-ID`

(You may need `gnupg-curl` installed.)

At that point, `xake --key YOUR-GPG-KEY-ID name reponame` will create
the remote repository `https://ximera.osu.edu/reponame.git` and you
will then be able to `git push ximera master` to store your work on
ximera.osu.edu.

## Theory of Operation

### Authorization via GPG

A key concern with Ximera is how to securely share grade data with
instructors; using GPG to authorize instructor-level access to Ximera
makes it possible for Ximera (through GPG) to send grade data securely
to the instructor.

### Using git to manage multiple versions of compiled assets

A challenge with Ximera is that, although the content may be changing,
we nevertheless want to be able to "go back in time" and see older
versions of the website.  This is importanf or data analysis (i.e., we
want to understand how an old version of an activity may be less
effective than a new version) and for the students (who may want to
see their previous work, even if the textbook authors update the
text).

As one would expect, the source code for the online activities is
stored in TeX files in a git repository.  Surprisingly, the compiled
assets are **also** stored in the git repository; if the source code
has commit hash SHA, then the publised files can be found in a tag
publications/SHA.  The command `xake frost` both creates this tag and
also performs a `git push` to send the commit to the server.

# Installing xake

```
mkdir -p ~/go/src/github.com/ximeraproject
export GOPATH=$HOME/go
cd ~/go/src/github.com/ximeraproject
git clone https://github.com/XimeraProject/xake.git
cd xake
go get .
go build .

```
1