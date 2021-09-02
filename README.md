# 19.2 短网址项目简介

有些浏览器中的地址（称为 URL）非常长且/或复杂，在网上有一些将他们转换成简短 URL 来使用的服务。我们的项目与此类似：它是具有 2 个功能的 *web 服务*（web service）：

## 添加 (Add)

给定一个较长的 URL，会将其转换成较短的版本，例如：
```
http://maps.google.com/maps?f=q&source=s_q&hl=en&geocode=&q=tokyo&sll=37.0625,-95.677068&sspn=68.684234,65.566406&ie=UTF8&hq=&hnear=Tokyo,+Japan&t=h&z=9
```
- (A) 转变为：`http://goto/UrcGq`
- (B) 并保存这对数据

## 重定向 (Redirect)

短网址被请求时，会把用户重定向到原始的长 URL。因此如果你在浏览器输入网址 (B)，会被重定向到页面 (A)。


https://github.com/unknwon/the-way-to-go_ZH_CN/blob/master/eBook/19.3.md