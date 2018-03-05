# 机器人技能Api

All URIs are relative to *http://127.0.0.1:8182/api/v1/robot/*

Method | HTTP request | Description
------------- | ------------- | -------------
[**functionsGet**](机器人技能Api.md#functionsGet) | **GET** /functions | 取得所有机器人技能开关状态




<a name="functionsGet"></a>
# **functionsGet**
> Componentsschemasfunction functionsGet()

取得所有机器人技能开关状态

### Example
```java
// Import classes:
//import io.swagger.client.ApiException;
//import io.swagger.client.api.机器人技能Api;



机器人技能Api apiInstance = new 机器人技能Api();

try {
    Componentsschemasfunction result = apiInstance.functionsGet();
    System.out.println(result);
} catch (ApiException e) {
    System.err.println("Exception when calling 机器人技能Api#functionsGet");
    e.printStackTrace();
}
```

### Parameters
This endpoint does not need any parameter.


### Return type

[**Componentsschemasfunction**](Componentsschemasfunction.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json



