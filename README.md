# Xake

[![Tag](https://img.shields.io/github/tag/XimeraProject/xake.svg?style=flat-square)](https://github.com/XimeraProject/xake/tags)
[![License: GPL v3](https://img.shields.io/badge/license-GPL%20v3-blue.svg?style=flat-square)](https://github.com/XimeraProject/xake/blob/master/LICENSE.md)
[![Build Status](https://travis-ci.org/XimeraProject/xake.svg?branch=master)](https://travis-ci.org/XimeraProject/xake)

Xake is the build tool for Ximera.  It is used to manage the
conversion of TeX files into .html files, and the publication of the
resulting .html files.  You may be interested in some of the
underlying [design principles](./docs/theory.md).

The easiest way to [install xake](./docs/install.md) is with a package manager.

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

## Using xake

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

At that point, `xake --key YOUR-GPG-KEY-ID reponame` will create
the remote repository `https://ximera.osu.edu/reponame.git` and you
will then be able to `git push ximera master` to store your work on
ximera.osu.edu. The name of your repository should contain only alphanumeric characters for now.

