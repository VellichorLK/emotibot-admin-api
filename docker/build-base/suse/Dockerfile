FROM docker-reg.emotibot.com.cn:55688/base/sles12sp3:2.0.2

ARG VERSION
RUN zypper ar http://download.opensuse.org/tumbleweed/repo/oss/ oss \
    && zypper --gpg-auto-import-keys ref \
    && zypper -n in --force-resolution \
    unzip \
    wget \
    vim \
    curl \
    glibc-locale \
    go${VERSION} \
    git \
    && rm -rf /var/lib/apt/lists/*

ENV LANG C.UTF-8
