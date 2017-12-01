# spaco-storage

storage# cosmos
# cosmos

## Task list
* data structure
* super client
* super node
* front node


## HTTP API

## calcuate_storage_fee
### try to calcute fee for storage file
```sh
URI: /cosmos/calcuate_storage_fee
Method: PUT
Args: 
    filesize: file data
    expire: expire date
    qos: quality of service 99%, 99.9%,99.99%,99.999%,99.9999% 
```

## announce_file
announce that I have a file

## file_put

### put file to cosmos filesystem

```sh
URI: /cosmos/file_put
Method: PUT
Args: 
    filename: filename(option)
    data: file data
    expire：expire date（0） means never expire(option)
    fee: 
```

example:

```sh
curl http://127.0.0.1:6420/cosmos/file_put
```

result:

```json

```

文件上传


### file_get
