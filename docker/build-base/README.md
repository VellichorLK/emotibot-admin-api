# Build-base

使用一個統一的Base image來做Go build 的底層image，
減少上層docker 需要每次重新pull 或是把第三方套件放在vendor 的 
並且理論上可以保證這個版本不會使不同版本互斥

## How we build

`./build.sh <sub-folder>`

sub-folder為不同的，參考章節[different OS for the build base](##different-os-for-the-build-base)
產生的docker image 格式為 sha1sum 前八碼-subFolderName
並產生一個VERSION檔案可供查看最新的版本號

## different OS for the build base

考慮到未來build-base彈性的關係，build base 本身可基於不同的 Image 或是 OS 來產生build base。
若需要擴充只要保證，資料夾內有一個 DOCKER_IMAGE 檔案供上層 build-base 使用，並且滿足

1. 已安裝golang
2. 已安裝git(供go get 使用)

目前已有的清單:

1. alpine(golang:1.9-alpine改編, general使用)
2. suse(for 民生專案?)