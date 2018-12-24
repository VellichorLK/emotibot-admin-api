# csv to ccqa import format

It is a simple util tool to transform a specify csv format to ccqa format for qi-controller.  
This is a specify case usage, which need some manual steps not write down on this page.

csv format header(optional):
    原数据,标签
Content columns:
    Column 1: training text, need a prefix char '0' -- 客服 or '1' -- 客戶.
    Column 2: intent tag, need a prefix char the same as col 1.
