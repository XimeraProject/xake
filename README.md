
# Xake

## Docker
- Builden: `docker build . -f Dockerfile -t set-registry.repo.icts.kuleuven.be/dsb/xake:latest`
- Updaten in docker-registry: `docker push set-registry.repo.icts.kuleuven.be/dsb/xake:latest`
- Runnen: `docker run --rm -v <absoluut_pad_zomercursus_repo>:/code set-registry.repo.icts.kuleuven.be/dsb/xake:latest xake -v`
- Runnen via bat file: 
    - Je kan rechtstreeks `xake compile` etc uitvoeren
    - Maak `xake.bat` file met volgende inhoud:
>>>
    @echo off    
    docker run --rm -v %cd%:/code set-registry.repo.icts.kuleuven.be/dsb/xake:latest xake -v %*
>>>

## Changelog 1/2023, v1.3.0: rework Dockerfile
- previous version did not compile (go version issues)
- include golang build into final image (so that hopefully minor further re-compiles might be possible INSIDE that image ...)
- use texlive-full prom debian packages  (might not be very fortunate ...)
- get rid of sage for now (image is already huge; did not work anyway)

## Nieuwe versie maken
Zie [readme van Ximera-server](https://gitlab.mech.kuleuven.be/monitoraat-wet/Ximera-server#nieuwe-versie-online-zetten)
Dbm kan gebruikt worden of handmatig.
Na het pushen, moet de `.gitlab-ci` file worden gewijzigd om de nieuwste versie te gebruiken.

## Uitleg originele repo

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

At that point, `xake --key YOUR-GPG-KEY-ID name REPONAME` will create
the remote repository `https://ximera.osu.edu/REPONAME.git` and you
will then be able to `git push ximera master` to store your work on
ximera.osu.edu. The name of your repository should contain only alphanumeric characters for now.

