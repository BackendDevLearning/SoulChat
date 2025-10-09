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