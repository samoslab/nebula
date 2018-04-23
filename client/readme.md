Index
-----

| Route                                                                                      | HTTP verb |
| ------------------------------------------------------------------------------------------ | --------- |
| [/store/list](#storelist-post)                             | POST      |
| [/store/register](#storeregister-post)                                   | POST      |
| [/store/verifyemail](#storeverifyemail-post)                             | POST      |
| [/store/resendemail](#storeresendemail-post)                             | POST      |
| [/store/upload](#storeupload-post)                                   | POST      |
| [/store/download](#storedownload-post)                             | POST      |
| [/store/remove](#storeremove-post)                             | POST      |


## 1. store register
```
URI:/store/register
Method:POST
Request Body: {
   email:string
   }
```

## 2. store verify email

```
URI:/store/verifyemail
Method:POST
Request Body: {
   code:string
   }
```

## 3. store create folder

```
URI:/store/mkfolder
Method: POST
Request Body: {
  "parent":"/"
  "":["abc","tmp"]
  "interactive":bool
}
```

## 4. store upload file

```
URI:/store/upload
Method: POST
Request Body: {
  "parent":/tmp
  "filename":"/tmp/abc.txt"
  "interactive":true
  "newversion" :false
  }
```

## 5. store list files

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

## store download files

```
URI:/store/download
Method: POST
Request Body: {
  filehash:string
  filesize:uint64
  filename:string
  folder:bool
  }
```

## 7. storeremove files

```
URI:/store/remove
Method: POST
Request Body: {
   filepath:string
   folder:bool
   recursion:bool
   }
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
