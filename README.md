# go-obfuscated-id
A value object used to identify domain objects. Domain entities only know themselves and can never cross it  Own object bounds to find out if the ID it generates is actually unique.  Once the new entity is saved as a record in the database, it will get an ID. That is, entities have no identity until they are persisted
