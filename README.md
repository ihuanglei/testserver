# testserver

测试服务器

使用js脚本编写配置文件

```javascript

// 支持方法:
// getHost() 返回当前请求主机信息
// getMethod() 返回当前请求方式
// getUri() 返回当前uri
// getQuery() 请求参数
// getForm() form表单提交参数
// getBody() 请求中的body

// 输出:
// 设置 result 值


// test url:
// http://127.0.0.1:7788
// http://127.0.0.1:7788/user/login?username=test&password=test

// 设置路由
var route = {
  '/': 'index',
  '/user/login': 'login'
}

function index() {
  var env = new Object
  env.host = getHost()
  env.method = getMethod()
  env.uri = getUri()
  env.query = getQuery()
  result = env
}

function login() {
  var username = getQuery().username || ""
  var password = getQuery().password || ""
  var user = new Object
  user.username = username
  user.password = password
  result = user
}


/**
 * app run
 */

var App = {
  run: function () {
    var uri = getUri()
    var url = uri.split('?')[0]
    var method = route[url]
    if (!method) {
      result = 'route not found'
      return
    }
    eval(method + '()')
  }
}

App.run()

```
