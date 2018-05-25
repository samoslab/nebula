Index
-----

| Route                                                                                      | HTTP verb |
| ------------------------------------------------------------------------------------------ | --------- |
| [/store/list](#storelist-post)                             | POST      |
| [/store/register](#storeregister-post)                                   | POST      |
| [/store/verifyemail](#storeverifyemail-post)                             | POST      |
| [/store/resendemail](#storeresendemail-post)                             | POST      |
| [/store/upload](#storeupload-post)                                   | POST      |
| [/store/uploaddir](#storeuploaddir-post)                                   | POST      |
| [/store/download](#storedownload-post)                             | POST      |
| [/store/downloaddir](#storedownloaddir-post)                             | POST      |
| [/store/remove](#storeremove-post)                             | POST      |
| [/store/progress](#storeprogress-post)                             | POST      |
| [/order/packages](#orderpackages-get)                             | GET |
| [/order/package/get](#orderpackageget-get)                             | GET |
| [/order/package/buy](#orderpackagebuy-post)                             | POST|
| [/order/all](#orderall-get)                             | GET |
| [/order/getinfo](#ordergetinfo-get)                             | GET |
| [/usage/amount](#usageamount-get)                             | GET |


统一说明 返回json object结构统一为： 成功：{"code":0, "data":object} 失败：{"code":1,"errmsg":"errmsg","data":object}  

  
## /store/register [POST]
```
URI:/store/register
Method:POST
Request Body: {
   email:string
   resend:bool (default false)
   }
```

## /store/verifyemail [POST]

```
URI:/store/verifyemail
Method:POST
Request Body: {
   code:string
   }
```

## /store/folder/add [POST]

```
URI:/store/folder/add
Method: POST
Request Body: {
  "parent":"/"
  "":["abc","tmp"]
  "interactive":bool
}
```

## /store/upload [POST]

```
URI:/store/upload
Method: POST
Request Body: {
  "filename":"/tmp/abc.txt"
  "interactive":true
  "newversion" :false
  }
```

## /store/uploaddir [POST]

```
URI:/store/uploaddir
Method: POST
Request Body: {
  "parent":/tmp
  }
```

## /store/list [POST]

```

URI:/store/list
Method: get
Request Body: {
  "path":"/tmp"
  "pagesize":10
  "pagenum":1
  "sorttype":name|size|modtime
  "ascorder":true
  }

```

Example

```
curl -X POST -H "Content-Type:application/json" -d '{"path":"/tmp/ok", "pagesize":10, "pagenum":1, "sorttype":"name", "ascorder":true}' http://127.0.0.1:7788/store/list
{
    "errmsg": "",
    "code": 0,
    "Data": [
        {
            "id": "f844e3f3-97a5-4da3-989e-ef354c8f4426",
            "filesize": 45382461,
            "filename": "/tmp/ok/testfile.big",
            "filehash": "8839307ab1fa4e37498136ddf47107058e33ecd5",
            "folder": false
        },
        {
            "id": "aa84ec51-c52c-41bf-bb65-8a28b6c8a57b",
            "filesize": 90764994,
            "filename": "/tmp/ok/erasure.12",
            "filehash": "7d5d901257ca0ac2fc170ade09f17524d195c6e8",
            "folder": false
        }
   ]
}

```

## /store/download [POST]

filehash and filehash is from /store/list result
download directory if parent isn't empty but others is empty , or download filename
```
URI:/store/download
Method: POST
Request Body: {
  filehash:string
  filesize:uint64
  filename:string
  }
```

## /store/downloaddir [POST]

```
URI:/store/downloaddir
Method: POST
Request Body: {
  parent:string
  }
```

## /store/remove [POST]

```
URI:/store/remove post
Method: POST
Request Body: {
   target:string
   folder:bool
   recursion:bool
   }
```

## /store/progress [POST]

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
curl -X POST -H "Content-Type:application/json" -d '{"files":[]}' http://127.0.0.1:7788/store/progress
{
    "errmsg": "",
    "code": 0,
    "Data": {
        "/tmp/abc/ipip.big1": 0.66
    }
}
```

## /order/packages [GET]

returns all packages

```
URI:/order/packages get
Method: GET
Args: None

```
## /order/package/get [GET]

returns one package
```
URI:/order/package/get get
Method: GET
Args: id

```

## /order/package/buy [POST]

buy package
```
URI:/order/package/buy POST
Method: POST
Request Body: {
   id:int,
   canceled:bool (default false),
   quanlity:int,
   }

```

## /order/all [GET]

returns all orders belong to you 
```
URI:/order/all get
Method: GET
Args: None

```

## /order/getinfo [GET]

returns all orders belong to you 
```
URI:/order/getinfo get
Method: GET
Args: orderid(string)

```
## /usage/amount [GET]

returns usage amount about order
```
URI:/usage/amount get
Method: GET
Args: None

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
