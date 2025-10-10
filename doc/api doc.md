# SoulChat API 文档

## 1. 概览

| 接口名称           | 方法 | 路径                                     | 功能说明           |
| ------------------ | ---- | ---------------------------------------- | ------------------ |
| Register           | POST | `/api/users`                             | 用户注册           |
| Login              | POST | `/api/users/login`                       | 用户登录           |
| UpdateUserPassword | POST | `/api/users/updatePassword`              | 更新用户密码       |
| UpdateUserInfo     | PUT  | `/api/users/updateUserInfo`              | 更新用户信息       |
| GetProfile         | GET  | `/api/profiles/{user_id}`                | 获取用户主页资料   |
| FollowUser         | POST | `/api/profiles/{target_id}/follow`       | 关注用户           |
| UnfollowUser       | POST | `/api/profiles/{target_id}/unfollow`     | 取消关注           |
| GetRelationship    | GET  | `/api/profiles/{target_id}/relationship` | 获取与他人关系状态 |
| CanAddFriend       | POST | `/api/profiles/{target_id}/canAddFriend` | 判断能否添加好友   |
| GetMessages        | GET  | `/api/chat`                              | 获取消息记录       |

## 2.接口详细说明

### 2.1 用户注册

#### **URL:** `/api/users`

#### **Method:** `POST`

#### Request Body

```json
{
  "username": "string",
  "phone": "string",
  "password": "string"
}
```

#### Response

```json
{
  "code": 0,
  "res": {
    "code": 200,
    "reason": "OK",
    "msg": "success"
  },
  "token": "string"
}
```
### 2.2 用户登录

#### **URL:** `/api/users/login`

#### **Method:** `/api/users/login`

#### Request Body

```json
{
  "phone": "string",
  "password": "string"
}
```

#### Response

```json
{
  "code": 0,
  "res": {
    "code": 200,
    "reason": "OK",
    "msg": "success"
  },
  "token": "string"
}

```
### 2.3 更新用户密码

#### **URL:** `/api/users/updatePassword`

#### **Method:** `POST`

#### Request Body

```json
{
  "phone": "string",
  "old_password": "string",
  "new_password": "string"
}
```

#### Response

```json
{
  "code": 0,
  "res": {
    "code": 200,
    "reason": "OK",
    "msg": "success"
  }
}
```

### 2.4 更新用户信息

#### **URL:** `/api/users/updateUserInfo`

#### **Method:** `PUT`

#### Request Body

```json
{
  "username": "string",
  "gender": 0,
  "birthday": "string",
  "bio": "string",
  "head_image": "string",
  "cover_image": "string"
}
```

#### Response

```json
{
  "code": 0,
  "res": {
    "code": 200,
    "reason": "OK",
    "msg": "success"
  }
}
```

### 2.5 获取用户主页资料

#### **URL:** `/api/profiles/{user_id}`

#### **Method:** `GET`

#### **Path Parameters:**

- user_id：string

#### Response

```json
{
  "code": 0,
  "res": {
    "code": 200,
    "reason": "OK",
    "msg": "success"
  },
  "data": {
    "user_id": 0,
    "tags": "string",
    "follow_count": 0,
    "fan_count": 0,
    "view_count": 0,
    "note_count": 0,
    "received_like_count": 0,
    "collected_count": 0,
    "comment_count": 0,
    "last_login_ip": "string",
    "last_active": "string",
    "status": "string"
  }
}
```

### 2.6 关注用户

#### **URL:** `/api/profiles/{target_id}/follow`

#### **Method:** `POST`

#### **Path Parameters:**

- target_id：string

#### Response

```json
{
  "code": 0,
  "res": {
    "code": 200,
    "reason": "OK",
    "msg": "success"
  },
  "data": {
    "self_id": 0,
    "follow_count": 0,
    "target_id": 0,
    "fan_count": 0
  }
}
```

### 2.7 取消关注

#### **URL:** `/api/profiles/{target_id}/unfollow`

#### **Method:** `POST`

#### **Path Parameters:**

- target_id：string

#### Response

```json
{
  "code": 0,
  "res": {
    "code": 200,
    "reason": "OK",
    "msg": "success"
  },
  "data": {
    "self_id": 0,
    "follow_count": 0,
    "target_id": 0,
    "fan_count": 0
  }
}
```

### 2.8 获取与他人关系状态

#### **URL:** `/api/profiles/{target_id}/relationship`

#### **Method:** `GET`

#### **Path Parameters:**

- target_id：string

#### Response

```json
{
  "code": 0,
  "res": {
    "code": 200,
    "reason": "OK",
    "msg": "success"
  },
  "data": {
    "is_following": true,
    "is_followed_by": true,
    "is_mutual": true,
    "is_blocked": true,
    "is_friend": true,
  }
}
```

### 2.9 判断能否添加好友

#### **URL:** `/api/profiles/{target_id}/canAddFriend`

#### **Method:** `POST`

#### **Path Parameters:**

- target_id：string

#### Request Body

```json
{
}
```

#### Response

```json
{
  "code": 0,
  "res": {
    "code": 200,
    "reason": "OK",
    "msg": "success"
  },
  "data": {
  }
}
```

### 2.10 获取消息记录

#### **URL:** `/api/chat`

#### **Method:** `GET`

#### **Query Parameters:**

#### Response

```json
{
  "code": 0,
  "res": {
    "code": 200,
    "reason": "OK",
    "msg": "success"
  }
}
```
