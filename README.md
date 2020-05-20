<h1 align="center">
  Git LFS WebDAV
  <br>
  <br>
</h1>

<h4 align="center">Store LFS files on any WebDAV server without a proper LFS backend.</h4>

<p align="center">
  <a href="https://github.com/mpotthoff/git-lfs-webdav/releases"><img src="https://img.shields.io/github/release/mpotthoff/git-lfs-webdav.svg?logo=github&style=flat-square" alt="GitHub Release"></a>
  <a href="https://github.com/mpotthoff/git-lfs-webdav/blob/master/LICENSE"><img src="https://img.shields.io/github/license/mpotthoff/git-lfs-webdav.svg?style=flat-square" alt="License"></a>
</p>

A [Custom Transfer
Agent](https://github.com/git-lfs/git-lfs/blob/master/docs/custom-transfers.md)
for [Git LFS](https://git-lfs.github.com/) which allows you to use a remote WebDAV
folder as the backend for your LFS files.

## How to use

For security reasons Git doesn't allow to configure a standalone transfer agent that will
be used for every clone automatically. The local repository needs to be configured after
every clone instead. So when cloning a repository Git won't know where to find the LFS files
and fail. There is really no way around this, but to make things easy I have hidden most of
the required configuration inside the `git-lfs-webdav init` command.

### Configure a new repository

* Configure your repository as usual with Git and Git LFS
* You will need the `git-lfs-webdav` executable
  * Either include it in your repository so that users won't have to install it
    (e.g. `.lfs/git-lfs-webdav-[platform]`)
  * or put it on `PATH` yourself (your users will have to do the same)
* Initialize WebDAV using
  * `./.lfs/git-lfs-webdav-[platform] init https://your/webdav/folder/` (if included)
  * or `git-lfs-webdav init https://your/webdav/folder/`
* Commit the created `.lfsconfig` (and the binaries if you have included them)
* Push everything as usual

### Clone an existing repository

* Clone the repository as usual using `git clone <url>`
  * This will work for the normal files but it will fail with `Error downloading object` for the LFS files
* Enter your cloned repository with `cd <folder>`
* You will need the `git-lfs-webdav` executable
  * Either it is included in your repository
    (e.g. `.lfs/git-lfs-webdav-[platform]`)
  * or you have to put it on `PATH` yourself
* Initialize WebDAV using
  * `./.lfs/git-lfs-webdav-[platform] init` (if included)
  * or `git-lfs-webdav init`
* As instructed by the initialization run `git reset --hard master` to fix your LFS files.

## Troubleshooting

### Authorize 401 Error

`git-lfs-webdav` will use the credential manager of git. Ensure that there is no leftover
entry in the configured credential backend (e.g. the Windows Credential Manager on Windows).

You should normally be prompted to enter a username and password by the configured
credential backend. If this does not work (for whatever reason) you can always configure
the login credentials using
  * `./.lfs/git-lfs-webdav-[platform] login` (if included)
  * or `git-lfs-webdav login`

The entered credentials will be stored in cleartext in the `.git/config` file of your local
repository (to prevent them from being committed accidentally).