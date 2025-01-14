# api-binds

http api绑定， 用于生成简单接口的序列化代码。

##### Start:

1. 实现handler

   ```go
   //go:generate api-binds -name handler -kind Gin
   type handler struct{
       
   }
   
   func (h *handler)PostRegist(req *types.RegistReq, resp *types.RegistResp)(code int, err error){
       // db.First ...
   }
   func (h *handler)PostLogin(req *types.LoginReq, resp *types.LoginResp)(code int, err error){
        // db.Find ...
   }
   func (c *handler)GetUserinfo(req *types.UserinfoReq, resp *types.UserinfoResp)(code int, err error){
       // db.Where ...
   }
   ```

   :warning: handler的方法必须以Get、Post开头, 表示接口对应的请求方法， 生成代码只会将Get的query参数、Post的body参数进行bind

   :warning: handler的入参必须是2个，依次表示req和resp，返回参数code为http status code, err 不为nil将设置为`http.StatusBadRequest`

2. 安装api-binds

   ```shell
   go install github.com/lysShub/api-binds@latest
   ```

3. 生成代码

   运行 `go generate ./...` , 可以看到有生成代码：

   ```go
   ```

   

   

##### TODO:

- 添加测试
- 支持Std
- 支持其他http请求方法、及Any
- 