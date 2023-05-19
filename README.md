# go-obfuscated-id

## Go Obfuscated ID
* ID is a value object that use for identify a domain object.
The domain entity is only aware of itself, and can never reach across its
own object boundaries to find out if an ID it has generated is actually unique.
Once a new entity gets persisted as a record in the database it will get an ID.
That is, the entity has no identity until it has been persisted.<br>

## USAGE



request model
```go
import (
    obfuscated "github.com/19byte/goobfuscated"
)

// RequestModel http request parameters
type RequestModel struct {
    ID   obfuscated.ID `json:"id" query:"id"`
    Name string        `json:"name" query:"name"`
}

// DBModel database query model
type DBModel struct {
    ID obfuscated.ID   obfuscated.ID `json:"id"`
    Name string        `json:"name"`
}

func GetModel(r *http.Request,w http.ResponseWriter)error  {
    // obfuscated.ID Implemented MarshalJSON & UnmarshalJSON interface
    m := &RequestModel{}
    if e := json.NewDecoder(r.Body).Decode(&m);e != nil {
        panic(e)
    }
    s := &DBModel{}
    db.QueryRow(r.Context(),`SELECT id,name FROM model`).Scan(&s.ID,&s.Name)
    return json.NewEncoder(w).Encode(s)
}

```