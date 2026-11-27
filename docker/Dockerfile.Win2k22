# syntax = docker/dockerfile:1
# escape=`

FROM mcr.microsoft.com/windows/servercore:ltsc2022

# Busybox Unicode https://github.com/rmyorston/busybox-w32
ARG BUSYBOX_VERSION=busybox-w64u-FRP-5467-g9376eebd8.exe `
    BUSYBOX_VERSION_SHA256=a78891d1067c6cd36c9849754d7be0402aae1bc977758635c27911fd7c824f6b

# Woodpecker plugin-git https://github.com/woodpecker-ci/plugin-git
ARG PLUGIN_VERSION=2.6.0 `
    PLUGIN_VERSION_SHA256=e09de6510c887127039f8fd26e8b72a30508c88b8e3e5524b442bc88ae610bfc

# Git installer https://github.com/git-for-windows/git
ARG GIT_VERSION=2.47.0 `
    GIT_VERSION_SHA256=b6ca6dcd5c818396faa57e06e10489aed3e16396317475ca8e88e30e4eb2e3c5

LABEL maintainer="Geco-iT Team <contact@geco-it.fr>" `
      name="geco-it/woodpecker-plugin-git" `
      vendor="Geco-iT"

SHELL ["cmd", "/S", "/C"]

# Install Busybox Unix Tools (https://github.com/rmyorston/busybox-w32)
RUN mkdir C:\bin && `
    curl -fSsLo /bin/busybox64u.exe https://frippery.org/files/busybox/%BUSYBOX_VERSION% && `
    /bin/busybox64u --install -s /bin && `
    /bin/echo "%BUSYBOX_VERSION_SHA256% /bin/busybox64u.exe" > SHA256SUM && `
    /bin/sha256sum -c SHA256SUM && `
    /bin/rm -f SHA256SUM

# Install Git
RUN curl -fSsLo git.tar.bz2 https://github.com/git-for-windows/git/releases/download/v%GIT_VERSION%.windows.2/Git-%GIT_VERSION%.2-64-bit.tar.bz2 && `
    /bin/echo "%GIT_VERSION_SHA256% git.tar.bz2" > SHA256SUM && `
    /bin/sha256sum -c SHA256SUM && `
    /bin/mkdir /git && `
    /bin/tar -xf git.tar.bz2 -C /git && `
    /bin/rm -f git.tar.bz2 SHA256SUM

# Install plugin
RUN curl -fSsLo /bin/plugin-git.exe https://github.com/woodpecker-ci/plugin-git/releases/download/%PLUGIN_VERSION%/windows-amd64_plugin-git.exe && `
    /bin/echo "%PLUGIN_VERSION_SHA256% /bin/plugin-git.exe" > SHA256SUM && `
    /bin/sha256sum -c SHA256SUM && `
    /bin/rm -f SHA256SUM

# Set System path
RUN setx /m PATH "C:\\git\\cmd;C:\\git\\mingw64\\bin;C:\\git\\usr\\bin;C:\\bin;%path%"

USER ContainerUser

# Install plugin
ENV GODEBUG=netdns=go

WORKDIR C:\woodpecker

SHELL ["bash.exe", "-c"]

ENTRYPOINT ["plugin-git.exe"]
