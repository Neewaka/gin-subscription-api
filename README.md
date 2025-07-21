## Build via docker compose

> docker compose build
> docker compose up


## API Routes

### Subscription

+ `/api/v1/subscription/{id}` - `GET` - returns single subscription.

Supported attributes:

| Attribute     |  In        | Type     | Required |
|:--------------|:-----------|:---------|:---------|
| `id`          |   path     | string   | Yes      | 

+ `/api/v1/subscription/` - `POST` - creates new subscription.

Supported attributes:

| Attribute     |  In        | Type     | Required |
|:--------------|:-----------|:---------|:---------|
| `service_name`          |   body     | string   | Yes      | 
| `price`          |   body     | int   | Yes      | 
| `user_id`          |   body     | int   | Yes      |
| `start_date`          |   body     | string   | Yes      |
| `end_date`          |   body     | string   | No      |


+ `/api/v1/subscription/` - `GET` - returns list of all subscriptions with filter

Supported attributes:

| Attribute     |  In        | Type     | Required |
|:--------------|:-----------|:---------|:---------|
| `user_id`          |   query     | int   | No      | 
| `service_name`          |   query     | string   | No      | 

+ `/api/v1/subscription/` - `PUT` - updates an existing subscription

Supported attributes:

| Attribute     |  In        | Type     | Required |
|:--------------|:-----------|:---------|:---------|
| `id`          |   path     | string   | Yes      |
| `service_name`          |   body     | string   | Yes      | 
| `price`          |   body     | int   | Yes      | 
| `user_id`          |   body     | int   | Yes      |
| `start_date`          |   body     | string   | Yes      |
| `end_date`          |   body     | string   | No      |

+ `/api/v1/subscription/` - `DELETE` - deletes an existing subscription

Supported attributes:

| Attribute     |  In        | Type     | Required |
|:--------------|:-----------|:---------|:---------|
| `id`          |   path     | string   | Yes      | 

+ `/api/v1/subscription/period-price/{period}` - `GET` - requests period of time in path, format "mm-yyyy:{mm-yyyy}", where right side might be ommited and autoreplaced with time.Now()

Supported attributes:

| Attribute     |  In        | Type     | Required |
|:--------------|:-----------|:---------|:---------|
| `period`          |   path     | string   | Yes      | 
| `user_id`          |   query     | int   | No      | 
| `service_name`          |   query     | string   | No      | 
