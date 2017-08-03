# Theory of Operation

## Authorization via GPG

A key concern with Ximera is how to securely share grade data with
instructors; using GPG to authorize instructor-level access to Ximera
makes it possible for Ximera (through GPG) to send grade data securely
to the instructor.

## Using git to manage multiple versions of compiled assets

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
has commit hash SHA, then the published files can be found in a tag
publications/SHA.  The command `xake frost` both creates this tag and
also performs a `git push` to send the commit to the server.


