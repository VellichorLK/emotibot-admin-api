# Docker Compose multistages build

拆分为以下两个 build stages:

## Build stage:

  包含两个 Docker build 的 stages:

    1. 复制所有需要的原始码并编译出 binary
    2. 只将所编译出来的 binary 复制至新的 layer 并生成 docker image，因此此 docker image 将只会包含所编译出来的 binary

## Pack stage:

  将 `build` 阶段所生成 docker image 中的 `binary` 及执行所需要的环境变数和设定档，复制进执行环境，并生成新的 docker image

## 资料夹结构:

- `./build/`: 包含 `Build stage` 所需的 shell script 及 docker files... etc
- `./pack/`: 包含 `Pack stage` 所需的 shell script 及 docker filers... etc
- `./build.sh`: 同时执行 `Build stage` 及 `Pack stage`
- `./run.sh`: 执行由 `Pack stage` 所生成的 docker image
- `./run.env`: 执行时所需的环境变数
- `./image_tags.sh`: 包含生成 docker image name, image tag, 及 container name 所需的环境变数
