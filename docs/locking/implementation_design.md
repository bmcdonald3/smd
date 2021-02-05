## Library (Go implementation)
```
Init(config)
	* Initialize connection to the server
	* setup structs (maps)
	* spawn up a sentinel thread that will sleep a second and then
	 BATCH renew locks. This will limit the thundering herd.
	* Idea is to have a buffer; we will sleep until the next closest
	 expiration - buffer.  Then we will renew everything within a 
	 window.  We will then sleep until the new next closest expiration - buffer. 
	
GetLock([]xname) []{xname, quit chan, public key, err error}
   * try to aquire the lock
   * register with the sentinel
CheckLock([]{xname,public key}) []{xname, quit chan,  err error}
   * validate the lock is still active
   * register with the sentinel
ReleaseLock() err error
	IF 
	   there is private key; will RELEASE the lock 
   FINALLY
		de-register with the sentinel
Shutdown()
   * will release any private keys
   * close any open channels
```


 * How many threads will this take? 1 for the senitinel
 * Can you return a nil chan? we think so
 
The advantage of having a sentinel is the batching; that way we can avoid a thunder herd request to the API.


## Pros and Cons 

### Separate Service

#### Pros:

1. we can scale it separately
2. we can develop it independently
3. there is flexibility if we ever add a policy 'agent' like entity

#### Cons:

1. we have to come up with EVERYTHING
2. we have to separately back up and recover
3. more overhead to the k8s system as a whole
3. we have to have a HSM synchronization mechanism (which will put load on HSM).


### Integrating with HSM:

#### Pros:

1. just need to alter table, add a new table then write the API + domain stuff
2. might be faster to develop
3. dont have to worry about backup/restore

#### Cons:

1. extra load on HSM
2. extra complexity in HSM
3. pretty much need to commit to the idea of no default rule (unless we want to make a new pod in  the container and keep growing HSM).

### Factor regardless of approach:

Regardless of which way we go

1. need to have deprecation, and converstion strategy for `hsm/v1/lock`.  
2. Its not clear how we will provide the CLI... unfortunately Swagger does NOT allow bodies on GET or DELETE, so everything is a POST.  And the way you specify custom verbs: `/path:custommethod` is not support in craycli, so we have to do `/path/custommethod`.

