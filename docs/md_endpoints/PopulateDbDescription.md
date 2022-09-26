Populate the database with the following data:

`two` users for authentication:

```json
[
  {
    "passphrase": "0b14d501a594442a01c6859541bcb3e8164d183d32937b851835442f69d5c94e",
    "username": "richard.sargon@meinermail.com",
    "name": "Richard Sargon"
  },
  {
    "passphrase": "6cf615d5bcaac778352a8f1f3360d23f02f34ec182e259897fd6ce485d7870d4",
    "username": "tom.carter@meinermail.com",
    "name": "Tom Carter"
  }
]
```

```text
password for 'richard.sargon@meinermail.com' user is: password1
password for 'tom.carter@meinermail.com' user is: password2
```

`ten` drones:


Model enum for a Drone:
```text
0 => Lightweight
1 => Middleweight
2 => Cruiserweight
3 => Heavyweight

```

State enum for a Drone:
```text
0 => IDLE
1 => LOADING
2 => LOADED
3 => DELIVERING
4 => DELIVERED
5 => RETURNING
```

generated drone collection:
```json
[
  {"serialNumber":"123e4567-e89b-12d3-a456-426614174001","model":2,...},
  {"serialNumber":"123e4567-e89b-12d3-a456-426614174002","model":1,...},
  {"serialNumber":"123e4567-e89b-12d3-a456-426614174003","model":3,...},
  {"serialNumber":"123e4567-e89b-12d3-a456-426614174004","model":1,...},
  {"serialNumber":"123e4567-e89b-12d3-a456-426614174005","model":3,...},
  {"serialNumber":"123e4567-e89b-12d3-a456-426614174006","model":0,...},
  {"serialNumber":"123e4567-e89b-12d3-a456-426614174007","model":2,...},
  {"serialNumber":"123e4567-e89b-12d3-a456-426614174008","model":3,...},
  {"serialNumber":"123e4567-e89b-12d3-a456-426614174009","model":0,...},
  {"serialNumber":"123e4567-e89b-12d3-a456-426614174010","model":0,...}
]
```

seven (7) medications:
```json
[
  {"name":"a random string","weight":115,"code":"a random code","image":"ZmFrZV9pbWFnZQ=="}, 
  {"name":"a random string","weight":10,"code":"a random code","image":"ZmFrZV9pbWFnZQ=="},
  {"name":"a random string","weight":210,"code":"a random code","image":"ZmFrZV9pbWFnZQ=="},
  {"name":"a random string","weight":34,"code":"a random code","image":"ZmFrZV9pbWFnZQ=="}
  ...
]
```