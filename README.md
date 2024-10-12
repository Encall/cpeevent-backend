# CPEEVO Backend side

Using Gin framework and MongoDB

## Create your virtual environment file

- Create the `.env` file in the main directory.
- You can look for template in `.env.template` file. Then copy your MongoDB connetion (e.g, your connection string is `mongodb://localhost:27017/`), paste the string inside the double-quote notation in your created `.env` file.

Your `.env` file will store the following information

```
MONGO_URI = "mongodb://localhost:27017/"
```

## Initialize the project

Initilize the project

```bash
go run main.go
```

Then you have to go to `localhost:8080`
