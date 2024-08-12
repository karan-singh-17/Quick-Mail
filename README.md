
# Quick Mail

Quick Mail is a Go-based API service designed for managing email campaigns, enabling users to create groups and send emails directly to them.  
It allows users to input recipients manually or via an online .csv file link. Similarly, message content can be entered manually or through an HTML template linked online.  
The service includes a custom-built two-factor authentication system that sends a verification code to the user's email, ensuring secure access.  
Additionally, Quick Mail can accept file paths for .csv or HTML files from the local machine when the server is run locally.

# API Reference

## User

### Register

```http
  POST /api/user/register
```

| Parameter | Type     | Description                |
| :-------- | :------- | :------------------------- |
| `email` | `string` | **Required**. Your email |
| `password` | `string` | **Required**. Your password |

###### If not already, first register to the service using a valid email and a password. 

### Verify

```http
  GET /api/user/verify/
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `-`      | `-` | - |

###### A mail will be sent to the email entered before to confirm and verify the user. Once the link in the email is accessed and verification is successful the user will be registered in the database.

### Login

```http
  POST /api/user/login
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `email` | `string` | **Required**. Your email |
| `password` | `string` | **Required**. Your password |

###### Enter your email and password with which you registered. Upon receiving the request, the API checks if the provided email exists in the database. If the email is found, the system generates a unique 6-digit code and sends it to the user's email address.

### Verify Login

```http
  POST /api/user/verify-login-code
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `code`      | `string` | **Required** Enter the code sent on email. |

###### This API checks the code and upon confirming it gives access to the user to use the other API's.


### Logout

```http
  POST /api/user/logout
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `-`      | `-` | - |

###### This API's securely get's the user logged out.

### Current User Information

```http
  GET /api/user/curr-user
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `-`      | `-` | - |

###### Gives the information regarding the current user.

## Groups

### Create Group

```http
  POST /api/group/create-group
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `name`         | `string`   | **Required** Name of the group |
| `recipients`   | `[]string` | **Required** Enter the email id's of the people who will be receiving the mail.|
| `subject`      | `string`   | **Required** Subject of the mail |
| `message`      | `string`   | **Required** Enter the message you want to send. |
| `csv_link`     | `string`   | **Optional** Can also enter the link to an online .csv file of email id's |
| `html_link`    | `string`   | **Optional** Can also enter the link to an online .html file to act as message|
| `csv_file_path`| `string`   | **Available On local Machine** Enter path to the .csv file |
| `html_path`    | `string`   | **Available On local Machine** Enter path to the .html file|

###### **Note:** From message , html_link and html_path only one can be sent. This also implies for csv_link and csv_path.

### Get Groups

```http
  GET /api/group/get-groups
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `-`      | `-` | - |

###### Shows all the groups of the current user with additional information.

### Execute Group

```http
  POST /api/group/execute-group
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `group_id`      | `string` | **Required** Enter the group_id of the group for execution|


###### Executes the group by sending the message to all the recipients of the group.

### Edit Group

```http
  PUT /api/group/edit-group
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `group_id`      | `string` | **Required** Enter the group_id of the group.|
| `name`         | `string`   | **Required** Name of the group |
| `recipients`   | `[]string` | **Required** Enter the email id's of the people who will be receiving the mail.|
| `subject`      | `string`   | **Required** Subject of the mail |
| `message`      | `string`   | **Required** Enter the message you want to send. |



###### Edit the group and add the new information.

### Delete Group

```http
  DELETE /api/group/delete-group
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `group_id`      | `string` | **Required** Enter the group_id of the group for deletion|

###### Delete's the group.

## Note

 #### To run this server locally make sure to generate a .env file with the following params
- ##### **from :-** Enter your email address from which mails would be sent. 
- ##### **password :-** Generate the App Password for your email and enter it.
- ##### **smtpHost :-** Enter your SMTP HOST name.(for gmail :- smtp.gmail.com)
- ##### **smtpPort :-** Enter your SMTP PORT used.(for gmail :- 587)
- ##### **database_url :-** Enter link to SQL database. Make sure to add (*?charset=utf8mb4&parseTime=True&loc=Local*) at the end of the link if not already entered.





