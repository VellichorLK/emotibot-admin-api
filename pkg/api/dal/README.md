# dal package

DAL 包支援與 BFOP 內的 dal module 進行API交互。
v1 目前保證向上支援到最新版本 BFOP V1.3 版本(2018.09.24)的 dal module

## Testing

dal 包內具有額外分開的daltest 包， 設計靈感來自於sql-mock，使用者可以直接使用 New 函式產生一組已被 mock 的 dal Client 以及 Mocker，透過 Mocker 可以直接注入測試中預期會出現行為。