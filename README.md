The project persists to a postgres DB

The following OS env variables must be set for the project to run:

DB_HOST
DB_PORT
DB_USER
DB_USER
DB_PASSWORD
DB_SSLMODE

The project exposes 2 endpoints:

1. localhost:8080/sign
   This accepts a body consiting of a userId and a string array of answers. eg:
   {
    "userId":"b1",
    "answers":["A","B"]
   }
   It also expects an auth header. A dummy one can be passed. It returns a json object with the signature. E,g:
   {
    "signature": "cca44d6c26b579e6b4a4cec27023451bafebae5d5dfa8244528a964d60acc2e5"
   }

3. localhost:8080/verify
   This accepts a body consisting of a userId and a signature. e.g:
   {
    "userId":"a1",
    "signature":"bb71dc78f8c3d5ebdb89a7fafb51ce8017e99d58930ab2885922d9e438f73e53"
   }
   It returns a json payload consting of answers and a timestamp. e.g:
   {
    "status": "OK",
    "answers": [
        "A",
        "B",
        "C"
    ],
    "timestamp": "2023-12-20T09:29:51.930705Z"
   }
