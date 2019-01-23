Index
-----

| Route                                                                                      | HTTP verb |
| ------------------------------------------------------------------------------------------ | --------- |
| [/api/v1/store/list](#apiv1storelist-post)                             | POST      |
| [/api/v1/store/register](#apiv1storeregister-post)                                   | POST      |
| [/api/v1/store/verifyemail](#apiv1storeverifyemail-post)                             | POST      |
| [/api/v1/store/resendemail](#apiv1storeresendemail-post)                             | POST      |
| [/api/v1/store/folder/add](#apiv1storefolderadd-post)                                   | POST      |
| [/api/v1/store/upload](#apiv1storeupload-post)                                   | POST      |
| [/api/v1/store/uploaddir](#apiv1storeuploaddir-post)                                   | POST      |
| [/api/v1/store/download](#apiv1storedownload-post)                             | POST      |
| [/api/v1/store/downloaddir](#apiv1storedownloaddir-post)                             | POST      |
| [/api/v1/store/remove](#apiv1storeremove-post)                             | POST      |
| [/api/v1/store/rename](#apiv1storerename-post)                             | POST      |
| [/api/v1/store/progress](#apiv1storeprogress-post)                             | POST      |
| [/api/v1/task/upload](#apiv1taskupload-post)                                   | POST      |
| [/api/v1/task/uploaddir](#apiv1taskuploaddir-post)                                   | POST      |
| [/api/v1/task/download](#apiv1taskdownload-post)                                   | POST      |
| [/api/v1/task/downloaddir](#apiv1taskdownloaddir-post)                                   | POST      |
| [/api/v1/task/status](#apiv1taskstatus-post)                                   | POST      |
| [/api/v1/task/delete](#apiv1taskdelete-post)                                   | POST      |
| [/api/v1/package/all](#apiv1packageall-get)                             | GET |
| [/api/v1/package](#apiv1package-get)                             | GET |
| [/api/v1/package/buy](#apiv1packagebuy-post)                             | POST|
| [/api/v1/package/discount](#apiv1packagediscount-get)                             | GET|
| [/api/v1/order/all](#apiv1orderall-get)                             | GET |
| [/api/v1/order/getinfo](#apiv1ordergetinfo-get)                             | GET |
| [/api/v1/order/recharge/address](#apiv1orderrechargeaddress-get)                             | GET|
| [/api/v1/order/pay](#apiv1orderpay-post)                             | POST |
| [/api/v1/order/remove](#apiv1orderremove-post)                             | POST |
| [/api/v1/usage/amount](#apiv1usageamount-get)                             | GET |
| [/api/v1/secret/encrypt](#apiv1secretencrypt-post)                             | POST |
| [/api/v1/secret/decrypt](#apiv1secretdecrypt-post)                             | POST |
| [/api/v1/service/status](#apiv1servicestatus-get)                             | GET |
| [/api/v1/service/root](#apiv1serviceroot-post)                             | POST|
| [/api/v1/service/filetype](#apiv1servicefiletype-get)                             | GET |
| [/api/v1/config/import](#apiv1configimport-post)                             | POST |
| [/api/v1/config/export](#apiv1configexport-get)                             | GET |
| [/api/v1/space/password](#apiv1spacepassword-post)                             | POST |
| [/api/v1/space/verify](#apiv1spaceverify-post)                             | POST |
| [/api/v1/space/status](#apiv1spacestatus-post)                             | POST |

统一说明 返回json object结构统一为： 成功：{"code":0, "data":object} 失败：{"code":1,"errmsg":"errmsg","data":object}  

  
## /api/v1/store/register [POST]
```
URI:/api/v1/store/register
Method:POST
Request Body: {
   email:string
   resend:bool (default false)
   }
```

Example

```
curl -X POST -H "Content-Type:application/json" -d '{"email":"16330@qq.com"}' http://127.0.0.1:7788/api/v1/store/register
{
    "errmsg": "",
    "code": 0,
    "Data": "ok"
}
```

## /api/v1/store/verifyemail [POST]

```
URI:/api/v1/store/verifyemail
Method:POST
Request Body: {
   code:string
   }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"code":"pf7v87ic"}' http://127.0.0.1:7788/api/v1/store/verifyemail

```

## /api/v1/store/folder/add [POST]

```
URI:/api/v1/store/folder/add
Method: POST
Request Body: {
  "parent":"/"
  "space_no":0,
  "":["abc","tmp"]
  "interactive":bool
}
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"parent":"/", "folders":["temp"], "interactive":false, "space_no":0}' http://127.0.0.1:7788/api/v1/store/folder/add
{
    "errmsg": "",
    "code": 0,
    "Data": true
}

```

## /api/v1/store/upload [POST]

```
URI:/api/v1/store/upload
Method: POST
Request Body: {
  "filename":"/tmp/abc.txt"
  "dest_dir": "/tmp"
  "interactive":true
  "newversion" :false
  "space_no":0
  "is_encrypt":false
  }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"space_no":0, "dest_dir":"/tmp", "filename":"/tmp/config-test.json", "interactive":false, "newversion":false, "is_encrypt":false}' http://127.0.0.1:7788/api/v1/store/upload
{
    "errmsg": "",
    "code": 0,
    "Data": "success"
}

```

## /api/v1/store/uploaddir [POST]

```
URI:/api/v1/store/uploaddir
Method: POST
Request Body: {
  "parent":/tmp
  "dest_dir": "/tmp"
  "space_no":0
  "is_encrypt":false
  }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"parent":"/tmp/bak", "dest_dir":"/tmp/bak", "space_no":0, "is_encrypt":false }' http://127.0.0.1:7788/api/v1/store/uploaddir
{
    "errmsg": "",
    "code": 0,
    "Data": "success"
}
```

## /api/v1/store/list [POST]

```

URI:/api/v1/store/list
Method: get
Request Body: {
  "path":"/tmp"
  "space_no":0
  "pagesize":10
  "pagenum":1
  "sorttype":name|size|modtime
  "ascorder":true
  }

```

Example

```
curl -X POST -H "Content-Type:application/json" -d '{"path":"/tmp/ok", "pagesize":10, "pagenum":1, "sorttype":"name", "ascorder":true,"space_no":0 }' http://127.0.0.1:7788/api/v1/store/list
{
    "errmsg": "",
    "code": 0,
    "Data": {
        "total":100,
        "files": [
        {
            "id": "f844e3f3-97a5-4da3-989e-ef354c8f4426",
            "filesize": 45382461,
            "filename": "/tmp/ok/testfile.big",
            "filehash": "8839307ab1fa4e37498136ddf47107058e33ecd5",
            "modtime": 10000,
            "filetype": "video",
            "extension": "avi",
            "folder": false

        },
        {
            "id": "aa84ec51-c52c-41bf-bb65-8a28b6c8a57b",
            "filesize": 90764994,
            "filename": "/tmp/ok/erasure.12",
            "filehash": "7d5d901257ca0ac2fc170ade09f17524d195c6e8",
            "modtime": 10000,
            "filetype": "audio",
            "extension": "mp3",
            "folder": false
        }
   ]
   }
}

```

## /api/v1/store/download [POST]

filehash and filehash is from /api/v1/store/list result
download directory if parent isn't empty but others is empty , or download filename
```
URI:/api/v1/store/download
Method: POST
Request Body: {
  filehash:string
  filesize:uint64
  filename:string
  space_no:uint32
  dest_dir:string
  }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"filehash":"732e7a7d3db77ffb6dde834c81d263dfd05922dc","filesize":68073855, "filename":"/tmp/ok/abc.txt", "space_no":0, "dest_dir": "/tmp/ok"}' http://127.0.0.1:7788/api/v1/store/download
{
    "errmsg": "too few shards given",
    "code": 1,
    "Data": ""
}

```

## /api/v1/store/downloaddir [POST]


```
URI:/api/v1/store/downloaddir
Method: POST
Request Body: {
  parent:string
  space_no:uint32
  dest_dir:string
  }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"parent":"/tmp/abc", "dest_dir":"/tmp/abc", "space_no":0}' http://127.0.0.1:7788/api/v1/store/downloaddir
{
    "errmsg": "",
    "code": 0,
    "Data": "success"
}
```

## /api/v1/store/remove [POST]

```
URI:/api/v1/store/remove post
Method: POST
Request Body: {
   target:string
   ispath:bool
   recursion:bool
   space_no:uint32
   }
```

Exmaple 

```
curl -X POST -H "Content-Type:application/json" -d '{"target":"62633239633363392d373462332d343961632d396633312d363731336331376433633334", "ispath":false, "recursion":false, "space_no":0 }' http://127.0.0.1:7788/api/v1/store/remove
{
    "errmsg": "",
    "code": 0,
    "Data": "success"
}
```

## /api/v1/store/rename [POST]

```
URI:/api/v1/store/rename post
Method: POST
Request Body: {
   src:string
   dest:string
   ispath:bool
   space_no:uint32
   }
```

Exmaple 

```
curl -X POST -H "Content-Type:application/json" -d '{"src":"62633239633363392d373462332d343961632d396633312d363731336331376433633334",  "dest":"newfile.txt", "ispath": false, space_no": 0}' http://127.0.0.1:7788/api/v1/store/rename
{
    "errmsg": "",
    "code": 0,
    "Data": "success"
}
```

## /api/v1/store/progress [POST]

returns all progress info if files is empty
```
URI:/store/porgress post
Method: POST
Request Body: {
   files:[]string
   }

```

Example

```
curl -X POST -H "Content-Type:application/json" -d '{"files":[]}' http://127.0.0.1:7788/api/v1/store/progress
{
    "errmsg": "",
    "code": 0,
    "Data": {
        "0@/tmp/def/test124/file1.4m": {
            "type": "UploadProgress",
            "rate": 1,
            "local": "/root/test124/file1.4m",
            "spaco_no": 0
        },
        "0@/tmp/def/test124/samllfile": {
            "type": "UploadProgress",
            "rate": 1,
            "local": "/root/test124/samllfile",
            "spaco_no": 0
        },
        "0@/tmp/def/test124/smallfile1": {
            "type": "UploadProgress",
            "rate": 1,
            "local": "/root/test124/smallfile1",
            "spaco_no": 0
        },
        "0@/tmp/download/app-1008.txt": {
            "type": "DownloadProgress",
            "rate": 1,
            "local": "d06334bcf4fbb554bc864c790d7c07ed48137c2f",
            "spaco_no": 0
        }
    }
}
```

## /api/v1/task/upload [POST]

async interface , task run at back-end

```
URI:/api/v1/task/upload
Method: POST
Request Body: {
  "filename":string
  "dest_dir":string
  "interactive":bool
  "newversion" :bool
  "space_no":int
  "is_encrypt":bool
  }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"space_no":0, "dest_dir":"/tmp", "filename":"/tmp/config-test.json", "interactive":false, "newversion":false, "is_encrypt":false}' http://127.0.0.1:7788/api/v1/task/upload
{
    "errmsg": "",
    "code": 0,
    "Data": "client-task:6"
}
```

## /api/v1/task/uploaddir [POST]

async interface , task run at back-end

```
URI:/api/v1/task/uploaddir
Method: POST
Request Body: {
  "parent":string
  "dest_dir": string
  "interactive":bool
  "newversion" :bool
  "space_no":int
  "is_encrypt":bool
  }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"space_no":0, "dest_dir":"/tmp", "parent":"/tmp", "interactive":false, "newversion":false, "is_encrypt":false}' http://127.0.0.1:7788/api/v1/task/uploaddir
{
    "errmsg": "",
    "code": 0,
    "Data": "client-task:7"
}
```

## /api/v1/task/download [POST]

filehash and filehash is from /api/v1/store/list result
download directory if parent isn't empty but others is empty , or download filename
async interface , task run at back-end
```
URI:/api/v1/task/download
Method: POST
Request Body: {
  filehash:string
  filesize:uint64
  filename:string
  space_no:uint32
  dest_dir:string
  }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"filehash":"732e7a7d3db77ffb6dde834c81d263dfd05922dc","filesize":68073855, "filename":"/tmp/ok/abc.txt", "space_no":0, "dest_dir": "/tmp/ok"}' http://127.0.0.1:7788/api/v1/task/download
{
    "errmsg": "client-task:12",
    "code": 0,
    "Data": ""
}

```

## /api/v1/task/downloaddir [POST]

async interface , task run at back-end

```
URI:/api/v1/task/downloaddir
Method: POST
Request Body: {
  parent:string
  space_no:uint32
  dest_dir:string
  }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"parent":"/tmp/abc", "dest_dir":"/tmp/abc", "space_no":0}' http://127.0.0.1:7788/api/v1/task/downloaddir
{
    "errmsg": "",
    "code": 0,
    "Data": "client-task:13"
}
```

## /api/v1/task/status [POST]

```
URI:/api/v1/task/status
Method: POST
Request Body: {
  "task_id":string
  }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"task_id":"client-task:7"}' http://127.0.0.1:7788/api/v1/task/status
{
    "errmsg": "",
    "code": 0,
    "Data": "done"
}
```

## /api/v1/task/delete [POST]

```
URI:/api/v1/task/delete
Method: POST
Request Body: {
  "task_id":string
  }
```

Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"task_id":"client-task:92"}' http://127.0.0.1:7788/api/v1/task/delete
{
    "errmsg": "",
    "code": 0,
    "Data": {
        "Key": "client-task:92",
        "Status": 0,
        "Task": {
            "type": "UploadFile",
            "payload": {
                "dest_dir": "/tmp",
                "filename": "/root/test125/file.bigbig",
                "interactive": false,
                "is_encrypt": true,
                "newversion": false,
                "space_no": 0
            }
        },
        "UpdatedAt": 1539525525,
        "Seq": 92,
        "Deleted": true,
        "Err": ""
    }
}
```
## /order/packages [GET]

returns all packages

```
URI:/api/v1/package/all get
Method: GET
Args: None

```

Example 

```
curl   http://127.0.0.1:7788/api/v1/package/all

{
    "errmsg": "",
    "code": 0,
    "Data": [
        {
            "id": "357096341043478529",
            "name": "month package",
            "price": 15000000,
            "volume": 1024,
            "netflow": 6144,
            "upNetflow": 3072,
            "downNetflow": 3072,
            "validDays": 30
        },
        {
            "id": "357096341154267137",
            "name": "season package",
            "price": 40000000,
            "volume": 1024,
            "netflow": 18432,
            "upNetflow": 9216,
            "downNetflow": 9216,
            "validDays": 90
        }
    ]
}
```

## /api/v1/package [GET]

returns one package
```
URI:/api/v1/package get
Method: GET
Args: id

```
Examples

curl   http://127.0.0.1:7788/api/v1/package?id=357096341154267137

```
{
    "errmsg": "",
    "code": 0,
    "Data": {
        "id": "357096341154267137",
        "name": "season package",
        "price": 40000000,
        "volume": 1024,
        "netflow": 18432,
        "upNetflow": 9216,
        "downNetflow": 9216,
        "validDays": 90
    }
}
```
## /api/v1/package/buy [POST]

buy package
```
URI:/api/v1/package/buy POST
Method: POST
Request Body: {
   id:string,
   canceled:bool (default false),
   quanlity:int,
   }

```

Example 

```
curl  -X POST  http://127.0.0.1:7788/api/v1/package/buy  -H "Content-Type:application/json" -d '{"id":"357615924202078209", "quanlity":1}'
{
    "errmsg": "",
    "code": 0,
    "Data": {
        "id": "31336164303736382d323330362d343736372d626131302d316332306265633131396332",
        "creation": 1529224680,
        "packageId": "357615924202078209",
        "package": {
            "id": "357615924202078209",
            "name": "basic package",
            "price": 15000000,
            "volume": 1024,
            "netflow": 6144,
            "upNetflow": 3072,
            "downNetflow": 3072,
            "validDays": 30
        },
        "quanlity": 1,
        "totalAmount": 15000000,
        "discount": 1,
        "volume": 1024,
        "netflow": 6144,
        "upNetflow": 3072,
        "downNetflow": 3072,
        "validDays": 30
    }
}
```

## /api/v1/package/discount [GET]

discount package

```
URI:/api/v1/package/discount GET
Method: GET
Args:
   id:string,
```

Example

```
curl http://127.0.0.1:7788/api/v1/package/discount?id=357096341043478529

```

## /api/v1/order/all [GET]

returns all orders belong to you 
```
URI:/api/v1/order/all get
Method: GET
Args: expired=[true|false] default is true

```

Example

```
curl   http://127.0.0.1:7788/api/v1/order/all?expired=true
{
    "errmsg": "",
    "code": 0,
    "Data": [
        {
            "id": "31336164303736382d323330362d343736372d626131302d316332306265633131396332",
            "creation": 1529224680,
            "packageId": "357615924202078209",
            "package": {
                "id": "357615924202078209",
                "name": "basic package",
                "price": 15000000,
                "volume": 1024,
                "netflow": 6144,
                "upNetflow": 3072,
                "downNetflow": 3072,
                "validDays": 30
            },
            "quanlity": 1,
            "totalAmount": 15000000,
            "discount": 1,
            "volume": 1024,
            "netflow": 6144,
            "upNetflow": 3072,
            "downNetflow": 3072,
            "validDays": 30
        }
    ]
}
```

## /api/v1/order/getinfo [GET]

returns all orders belong to you 
```
URI:/api/v1/order/getinfo get
Method: GET
Args: orderid(string)

```

Example

```
curl   http://127.0.0.1:7788/api/v1/order/getinfo?orderid=31336164303736382d323330362d343736372d626131302d316332306265633131396332
{
    "errmsg": "",
    "code": 0,
    "Data": {
        "id": "31336164303736382d323330362d343736372d626131302d316332306265633131396332",
        "creation": 1529224680,
        "packageId": "357615924202078209",
        "package": {
            "id": "357615924202078209",
            "name": "basic package",
            "price": 15000000,
            "volume": 1024,
            "netflow": 6144,
            "upNetflow": 3072,
            "downNetflow": 3072,
            "validDays": 30
        },
        "quanlity": 1,
        "totalAmount": 15000000,
        "discount": 1,
        "volume": 1024,
        "netflow": 6144,
        "upNetflow": 3072,
        "downNetflow": 3072,
        "validDays": 30
    }
}
```

## /api/v1/order/pay [POST]

pay order
```
URI:/api/v1/order/pay POST
Method: POST
Request Body: {
   order_id:string,
   }

```

Exmpale 

```
curl  -X POST  http://127.0.0.1:7788/api/v1/order/pay  -H "Content-Type:application/json" -d '{"order_id":"31336164303736382d323330362d343736372d626131302d316332306265633131396332"}'
{
    "errmsg": "",
    "code": 0,
    "Data": {}
}
```

## /api/v1/order/remove [POST]

remove order
```
URI:/api/v1/order/remove POST
Method: POST
Request Body: {
   order_id:string,
   }

```

## /api/v1/order/recharge/address [GET]

returns pay-address and balance
```
URI:/api/v1/order/recharge/address get
Method: GET
Args: None

```

Example

```
curl http://127.0.0.1:7788/api/v1/order/recharge/address
{
    "errmsg": "",
    "code": 0,
    "Data": {
        "address": "29deQdbg3GKBueNhdX7FNJyC23HXk9JwjdU",
        "balance": 0
    }
}
```

## /api/v1/usage/amount [GET]

returns usage amount about order
```
URI:/api/v1/usage/amount get
Method: GET
Args: None

```

Example 

```
curl http://127.0.0.1:7788/api/v1/usage/amount
{
    "errmsg": "",
    "code": 0,
    "Data": {
        "packageId": 357615924202078209,
        "volume": 1024,
        "netflow": 6144,
        "upNetflow": 3072,
        "downNetflow": 3072,
        "usageVolume": 1,
        "usageNetflow": 6,
        "usageUpNetflow": 12,
        "usageDownNetflow": 32,
        "endTime": 1531819508
    }
}
```

## /api/v1/secret/encrypt [POST]

encrypt file
```
URI:/api/v1/secret/encrypt POST
Method: POST
Request Body: {
   file:string,
   password:string,
   output_file:string,
   }

```

Exmpale 

```
```

## /api/v1/secret/decrypt [POST]

decrypt file
```
URI:/api/v1/secret/decrypt POST
Method: POST
Request Body: {
   file:string,
   password:string,
   output_file:string,
   }

```

Exmpale 

```
```

## /api/v1/service/status [GET]

returns service status
```
URI:/api/v1/service/status
Method: GET
Args: None

```

Example 

```
curl   http://127.0.0.1:7788/api/v1/service/status
{
    "status": true
}
```

## /api/v1/service/filetype [GET]

returns all supported file types
```
URI:/api/v1/service/filetype
Method: GET
Args: None

```

Example 

```
curl   http://127.0.0.1:7788/api/v1/service/filetype
{
  "application/epub+zip": {
        "type": "application",
        "sub_type": "epub+zip",
        "value": "application/epub+zip",
        "extension": "epub"
    },
    "application/font-sfnt": {
        "type": "application",
        "sub_type": "font-sfnt",
        "value": "application/font-sfnt",
        "extension": "ttf"
    }
}
```

## /api/v1/service/root [POST]

set root path
```
URI:/api/v1/service/root
Method: POST
Args: 
   root: string

```

## /api/v1/space/password [POST]

set space password 
```
URI:/api/v1/space/password
Method: POST
Args: 
   password: string
   space_no: uint32

```
Example 

```
curl -X POST -H "Content-Type:application/json" -d '{"password":"12345678abcdefg", "space_no":0 }' http://127.0.0.1:7788/api/v1/space/password
```

## /api/v1/space/verify [POST]

check space password correctness
```
URI:/api/v1/space/verify
Method: POST
Args: 
   password: string
   space_no: uint32

```

## /api/v1/space/status [POST]

check space psssword set or not 
```
URI:/api/v1/space/status
Method: POST
Args: 
   space_no: uint32

```
Example
```
curl -X POST -H "Content-Type:application/json" -d'{}' http://127.0.0.1:7788/api/v1/space/status
{
    "errmsg": "not buy any package order",
    "code": 401,
    "Data": ""
}
```

## /api/v1/config/import [POST]

import client config info
```
URI:/api/v1/config/import
Method: POST
Args: 
   filename : string

```

## /api/v1/config/export [GET]

export client config info
```
URI:/api/v1/config/export
Method: GET
Args: 

```

# websocket interface

port: 7799
uri: /message

消息格式：

```
1. {"type":"UploadFile","key":"0@/tmp/xab-1012","local":"/root/test124/xab-1012","space_no":0,"code":0,"error":""}
2. {"type":"UploadDir","key":"/root/test124","local":"/tmp/def","space_no":0,"code":0,"error":""}
3. "type":"DownloadFile","key":"0@/root/tmp/app-1008.txt","local":"/tmp/download/app-1008.txt","space_no":0,"code":0,"error":""}
4. {"type":"DownloadDir","key":"/root/test", "local":"/tmp","space_no":0,"code":0,"error":""}
5. {"type":"UploadProgress","key":"0@/tmp/xab-1012","local":"/root/test124/xab-1012","progress":1,"space_no":0,"code":0,"error":""}
6. {"type":"DownloadProgress","key":"0@/root/test124/CMakeLists.txt", "lcoal":"/local/CMakeLists.txt","progress":1,"space_no":0,"code":0,"error":""}
```

# specification

```
文件如果需要加密就先加密，然后计算sha1 hash，然后进行第一步，发送请求检查文件是否存在（CheckFileExist），注意如果文件size<=8k，文件数据随请求一起发送，metadata server判断如果文件已经存在，就返回done=tue，整个过程完成，不再执行第二步和第三歩

如果文件不存在，返回请求告诉使用纠删码还是多副本（根据文件大小判断），如果要使用纠删码还会返回分多少片多少个校验片，然后执行第二步（UploadFilePrepare）；如果使用多副本，还会返回应该使用多少个副本，以及用来存储这些副本的Provider，不需要执行第二步，直接跳到第三歩

如果使用纠删码，第二步发请求把按照纠删码计算获得的分片hash和片大小发送到服务器，服务器端返回用来存储分片的Provider，注意大部分Provider是对应某个hash的分片，但有少部分provider是作为备用的，可以存储多个分片其中任意一个

不管是使用纠删码还是多副本，数据存储到Provider完成后，再执行第三歩UploadFileDone把分区和分片信息传到服务器端，如果是多副本方式，分区和分片数都是1，存储的Provider(storeNodeId)是多个

纠删码上传
纠删码下载
```

