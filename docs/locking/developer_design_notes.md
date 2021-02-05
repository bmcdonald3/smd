# Locking Service Design 
## Developer Notes

### Goals of THIS working group:

 1. Define the boundary of the system
 2. define the high level capabilities and requirements
 3. scope primary work package
 4. define other services that MUST change to adapt
 5. Understand impact to v1.4 release and how to re-shuffle
 6. COME up with an execution plan; and HOW we can make this happen.  

### Development Tasks

 1. Create API specification (swagger)
 2. Create SQL model
 3. Cursory Exploration into technologies/packages that can help us
 4. Implement API -> Mk0
 5. Implement Agent -> Mk1 (we may defer implementation of this... so we would MANDATE that someone do this manually).
 6. Library -> so our go services can use it.
    7. other services have to integrate and change as well. 
 7. CLI -> This might be custom. Should use the Admin api? not the service api
 8. Investigate failure modes/ resiliency testing.  THIS thing HAS to work!  HIGHLY scalable; 


###Scenarios:

*These scenarios help us consider the system, what it does, what it could do, why it should do it. These scenarios are not a list of requirements but rather a shared dialogue tool to ground the engineers with similar viewpoints.*

  1. CAPMC receives a power off command, and locks the xname so a service like FAS cannot initiate a firmware update on the same xname at the same time. The opposite scenario is also valid:  FAS is running an update, so it locks an xname and keeps it locked (renews) until it is finished.  This prevents CAPMC from changing the power state to protect the update process. 
  2. IFS starts the process to update the GPU firmware of an xname. It locks the XNAME.  IFS informs BOS to reboot the xname into the inband firmware update image. BOS tells CAPMC to reboot the xname. CAPMC FAILS because IFS holds the lock.
  3. An admin wants to directly hit Redfish to accomplish task X, which cannot be done through a service.  The admin LOCKS the xname so no service can modify the xname. 
  4. A user tells CAPMC to power down all nodes (a wide open query).  The NCNs are 'closed' by default, and are therefore locked.  This prevents CAPMC from powering down the xnames of the NCNs
  5. an admin wants to facilitate updating the firmware of an NCN; they must specifically 'unlock' the NCN so that FAS can perform the update.
  6. The system admin needs to perform an emergency power off of an xname. They brake the lock; the lock has to be replaced before FAS can attempt a firmware update.  -> THIS is not for an NCN; that is probably a separate case.
  7. An admin can lock xnames (as part of their application pool) and unlock them; so that the xnames are protected from things like CAPMC
  
### General Discussion

The following are the discussion notes from Matt, Sean, Andrew while they hashed out the design and requirements.

**NOTE**: The following sections will give good insight to *what* we discussed and the evolution of the conversation.  It is VERY likely that there are things in these notes that do not align with the above requirements.  That is because we have undertaken an iterative development and design model, so things have been pretty fluid.  Care has been taken to make sure the requirements are as accurate as possible.  These notes are left to give good context to all that we considered. 
  
  1. Should locking a BLADE mean locking everything below it?
    2. We should make it hierarchical (optionally)... so everything below is locked. We will 
    3. if its the chassis; whats actually below it? just the chassis controller? 
    4. we will have the capability to do wildcard* xnames; like capmc -> the hierarchy only goes down.
    5. good question brought up by Michael about EPO and cooling groups. EPO a chassis and the cooling group takes the other chassis/cabs in the cooling group with it. 
  3. who can lock, unlock, how?  
     4. locks have expiration times and renewal times. 
     5. who can lock: services can LOCK, and an admin can lock. 
     6. who can unlock: services can unlock themselves; but they cannot unlock anything else.
     7. Admins can have forever locks; services MUST renew locks; services cannot have forever locks. renewable locks must be 0 < X > upper bound. You can only unlock, what you've locked.   WHen you take a lock, you get a token.  You can use that token to unlock and/or renew; its a single use token (like oauth).
  4. how does breaking work?
     5. IF it breaks its broken.  Renewal will fail; the xname must be 'restored/repaired/etc'.   
 6. What is the replace/repair process?
     7. an admin broke it (invalidate, some other phrase).
     8. capmc broke it
     9. LOCK broke it, because it WENT bad
     10. Allow them to go right to lock/unlock (depends on what they want). 
     11. CAPMC can break a lock; but no other service.  ONly an admin can repair a lock.
 7. How do we know about xnames as they come and go?
     8. we always keep a record of the xname; if it WAS an xname we keep it... if the component gets deleted, we mark it as deleted (but we arent actually removing the record); we would probably break the lock on it. 
 8. What about the PDUs?
     3. PDU's have two xnames:  the jaws (hostname), and the RTS xname.  The individual outlets DO have xnames.
     4. PDUs can be locked, and hierarchically; but we wont lock the PDUs for a BMC, just because they lock the BMC; just like we WONT lock the switch over top it. 	ANd we dont preemptively lock CDUs, etc, etc.

     8. So how do we protect PDUs?
        9. Have an admin LOCK the PDUs;
        10. we are going to rely on authz to make sure people cant just get there; then use a locking process to make sure an authorized user can accidentally turn the thing off. 	

     
 11. How do I share access?  How does FAS tell CAPMC, here is my token, you can borrow it?
   12. Couple of ways (maybe)
      1. Pass a token to the service (IFS pass to BOS, BOS pass to CAPMC)   THIS WILL require a BOS and CAPMC API change!  We should get their buy in!
         2.  the CAPMC request to locks would include the token (which would let them 'inherit' or share the lock? 
         3. So maybe someway we get a subordinate token.
      2. somehow  pre-register the service -> NO, not an option b/c services cannot register intent.
  3. How do we revoke the token?
      4. Every subordinate token can be invalidated by the parent. 

8. Can bulk operations return a bulk ID?
  9. NO! we spent a long time deliberating on this.  The 'advantage' of an ID to represent a set of tokens is nice on the surface, but the overwhelming complexities of set transformation, set identity are just too much.  You can ALWAYS use an ARRAY of tokens, but some other arbitrary id just has too many combinatorial complexities.   
     10. 1 ID could represent many tokens.  Can many subordinate tokens be generated and then represented by a new ID? Make a request for 100 xnames;  I get back 1 ID + 100 tokens. generate subordinate tokens: and ID, OR list of tokens.  we get back a NEW ID + 100 tokens. 
     12. Set membership {a,b,c,d,e}, set identity `asd32lfkjasdflas` ={1,4,s,g};  replace or add an element, and `asd32lfkjasdflas` is still `asd32lfkjasdflas`. 
  10. One thing we have considered is that the token would be a combination of xname:somestringofnumbers that way you can always see WHAT this thing is for. 
  11. One other thing we have considered is that an admin needs to generate tokens for xnames (b/c they are locked) that it could be > to a file; then the CLI could accept a file as input to not have to include a long list of tokens
12. What do you do for things that are already locked and the request to LOCK is sent?  What do you do for things that are unlocked, and the request to UNLOCK is sent?  
   13. We want to handle them the same way... the caller is giving us an ACTION, not an end state. IF you tell us to UNLOCK something already unlocked it will fail.  We will tell you why it failed.  This is very nuanced.  the Redfish standard requires that a power off to an already off device fails, B/C the ACTION failed. 
14. Why would an admin ever need to use a subordinate lock?
   2. the admin does not want to unlock an NCN, but rather pass the subordinate token to a service to act on its behalf (delegation). 
2. Could an admin have an expiring lock?
   3. No admin locks are forever (unless they unlock or break), but they can generate a token.  That token will expire.
4. Should there be an audit log that has an api w/ operational persistence for it?
   5. Fundamentally we believe that there should be an audit/event log, but that this should follow the normal Shasta pattern, which is logs + kubernetes + smf (etc).  While it may be advantageous to have a audit API this would be a deviation from the system administration paradigm.  It would also be a costly feature to implement; given that we must manage transactions and callback (versus just logging from the service as needed). 
6. What is the scaling model for this service?
   7. it needs to be able to handle several thousand requests per second, as it will be a commonly touched interface by large volume services (CAPMC, FAS, IFS).  At this point we do not think that it will be a full order of magnitude more performant; but it must be on par with how much we use HSM, or CAPMC.   
    
### Policy Discussion

 11. Should there be an agent that executes policies?  e.g. lock all NCNs when you see it. 
     12. No policy layer at the moment... we aren't sure that it wouldn't be more painful that beneficial. 
     13. good scenario: auto-locking components, so they DON'T auto power on.  However this conflicts with the NO touch discovery mechanism...
 14. BUT How do we protect the NCN's ?
     15. Should there be a policy (or config) that will LOCK NCN's automatically?
 16. How would we SELECT things to LOCK 
     17. XNAME, ROLE, DeviceType (compType?)
     18. How do we SET this?  config-map? API?  
         19. API + LoaderJob + config-map that is used by loader job
         20. imagine we BAKE the defaults into the SQL -> part of the database migration.
         21. If you want to override those; in the manifest, you would have key-values for the LOADER.  The Loader will take that config and apply it. 
 22. What does this policy application look like? only on NEW things? or GO DO IT NOW!; does it change state? will the policy apply at any future time?
     23. We need to decide if we honor it for the future? or do everything...
     24. After much deliberation we agree that it must be retroactive, because of uncertainty of when things we be 'found'.  
     25. We realize this is a fair amount of work, but the edge cases on timing are just too broad.
 25. Related question: if WE enforce a policy and LOCK something; HOW do we give an admin a token? dont they HAVE to break it?  If they UNLOCK the xname; will it ever again re-lock?  
     26. We decided that an admin can just generate a token for a locked xname; that token will auto expire after X mins, so its really just for passing down the chain. 

### Further discussion on the validity of tokens and the flow of the process.

So what happens if KEYs dont rotate? -> really the key rotation was really just to FORCE a service that it couldnt CACHE keys... because if it did it would either have to be REALLY clever or would lose access.   

Every CHAIN should have its own token; and every link in the chain should have its own token, but MAYBE we DONT force them to rotate.  The math to figure out WHO belongs to WHO was getting too weird

^ maybe to compensate there is a token + a key; the key rotates but the token doesnt.  that probably makes the book keeping easier

RESPONSE: not rotating is good; but the intent behind the issue was to make sure services didnt do bad things.  That is realistically too much to ask for locking to compensate for.

***

Who has to renew?
The parent? well they care the most, but what about when the PARENT is the admin?

The last most child?  well in the case of things like FAS; they might wait a few minutes before they try to do that... actually if something else is running thats a guarantee!

What is the purpose of renewal? it tells the service like FAS; you are still authorized to go do something.  It tell the Locking service, HEY im still here; and therefore probably tells the parent (still around)... But that might be bad -> because thats really a separate concern.  

RESPONSE: Locking doesnt exist for general purpose IPC.  NO ONE should EVER assume because a token has been released means that the process that called it, or the process called is finished; that is WAY beyond the scope of this!

***

IF there is NO renewal, how does it know when it shouldn't do something?  it could just say, "well I know I have the lock for x mins; so i will only check x mins."  But if we remove that requirement then the service REALLY needs to check before it does anything! 

RESPONSE:  Before the service does ANYTHING to actual hardware (eg a LITERAL redfish call or ipmi call) it MUST check the lock to see if it is allowed.

***

What about releasing? How do we know when everything is done?

Reference counts -> and passing a subordinate all the way down.  any level can invalidate below it.  -> but how do you make sure you don't OVER pop?
follow good hygiene and release your lock!
when the top level caller is done, it releases it ref; and then the token disappears. But how do we know who the very top is?

RESPONSE: If a service doesn't get a key, it knows it is the generator, so it must renew; else it depends on the generator to renew.

***

After literally hours of back and forth on subordinate tokens and token rotation, and token renewal we have simplified.
you get a KEY.
a key has two parts: public/private.  You use the public to tell other people to do stuff; but you have to use private to renew/release.

Admin keys dont ever expire, as a good practice they SHOULD release them when they are done using them.  An admin has to use the private key to release it; else someone CAN break it.   Ad admin uses the public key to get other things to do his bidding.

A service CAN get a key (pub/private)

We came to this position because we cannot completely compensate for human behavior, nor should we try.  This model does accomplish the primary objective of only allowing admins to issue commands on locked hardware, and only if they follow certain processes.  Futhermore this does a MUCH better job of ensuring that services WONT cut each other off.


## Rough - somewhat overly prescriptive requirments

### Behavioral  / Functional

 1. The system shall provide an interface that will:
   2. allow the state of xnames to be viewed. This information will include:
      3. lock state
      4. reservation state (includes expiration time)
      6. 'who' took the lock or reservation
   2. allow xnames to be locked.
      6. locking may only be done via the admin interface. Locking done by the admin interface will not produce a reservation, but a reservation mechanism shall be provided. 
      7. locks have no expiration time, but may be unlocked.
   2. allow xnames to be reserved.
     3. The act of reservation shall cause a deputy/reservation key pair to be generated.  
         4. The deputy key may be passed at the callers discretion to a delegate process.  Possession and utilization of deputy key allows the holder to use the reserved hardware for the intent of the original caller.   The keys should not be cached for future use; only to complete the request of the original caller. The delegate may pass the key on to other delegates as necessary.
         5. The reservation key may be used to release and renew the reservation.  The reservation key should not be shared with other processes. 
         6. A new reservation/deputy pair cannot be generated as long as the current reservation is in effect.
     3. any reservation held by a service may only be held for a limited duration *d* not to exceed 15 minutes and no less than 1 minute, and in increments of 1 minute.  The reservation may be:
         4. Released - which removed the reservation and allows another caller to request a reservation.
         5. Renewed - which allows the current reservation holder to renew their lease and extend the expiration time out *d* minutes.  
             6. A reservation may be renewed subsequent times, there is no limit to the number of time a reservation may be renewed  (extended). 
             7. No reservation may be extended more than one period out (e.g. 5 back to back calls to renew only extend the reservation a total length of 1 period until the current period elapses). 
             8. The entity that generated the deputy/reservation key MUST renew the lock, as they have the reservation key, no one else may renew or release the reservation.
         6. Expired - the caller does not renew the reservation, the reservation expires and is removed.  The previous reservation holder may not renew the reservation, but they may attempt to get a new reservation.  
     3. any reservation held by an admin will be held for an infinite duration (e.g. admin reservations do not automatically expire). The reservation may be:
         4. Released - which removed the reservation, but not the lock, and allows an admin to request a new reservation.
     5. an admin must hold the lock before they can generate a reservation
  
 	4. allow an admin or authorized service to break a lock and or break reservations. The purpose of this mechanism is to allow an admin to take control of an xname and prevent a service from locking or reserving the xname.   Additionally the purpose is to allow CAPMC to submit an Emergency Power Off (EPO).
 	   5. a broken lock cannot be unlocked or locked or reserved.  It must be repaired.  Once the lock is repaired it can be set into locked or unlocked state. Once the lock is repaired reservations can be taken against the xname.  
 	   6. No service shall be able to repair a lock; only an admin.
 	   7. a broken lock remains broken until it is repaired. No policy agent may compensate for this.
   8. allow  an admin to repair a broken lock
     9. the lock may be put back into locked or unlocked status. Reservations may now be taken against the xname. 
5.  log every request and event. Specifically:
   6. locking
   7. unlocking
   8. breaking
   9. repairing
   10. releasing
   10. renewing
   11. delegating


### Policy Agent 
#### Decision Point 
We need to flush out how much of an active policy agent we will have for initial release.   The tradeoffs are:

 * Nothing automatic -> admins must identify and control
 * fully automatic -> we have to develop an API to codify rules and then come up with the algorithm for applying them. This will take some time!
 * partially automatic  -> we bake in a rule like 'lock all NCNs' the first time you see them.  Downside is there is NO visibility to this, and no easy way to change it. Probably more harm than its worth.

#### Requirements   

1. allow for creation of locking policy.  A locking policy is a general rule that will be applied to a set of xnames.  The intent of a locking policy is to create sane defaults for certain classes of xnames in order to protect the system.  Examples include: locking the NCNs so that they can not be power controlled without first being unlocked (or having a delegate created). The locking policy shall:
     2. Allow an admin to create, read, update, delete policies.
     3. Allow an admin to specify xname lock state of: `locked`. The default state for any xname is `unlocked`.
     5. Allow an admin to specify a filter for xname selection to include:
         6. role
         7. xname
         8. wildcarded xname
         9. component type (or device type?) 
 
6. The system shall have an agent that will enforce locking policies
     7. Locking policies can be added that locks certain xnames according to filters that identify the xname as part of some set (role, deviceType, name, etc). 
     8. **RESOLVE:** should the lock status be uninitialized? or default to unlocked?
     9. The system shall continually scan for new xnames in HSM and apply policies.
     10. Policies shall be immediately enforced upon creation or or upon new conditions that match the criteria.
     10. **RESOLVE** can there be conflicting policies?  If the only position allowed is UNLOCKED then no; but if a policy can be LOCKED or UNLOCKED we might get into trouble. 
     11. **RESOLVE** will the policy layer try to reset anything after the lock is repaired? How would we protect the system from a bad cycle of break, repair, re-lock.  Really the question is can an ADMIN unlock without a token?
     12. **RESOLVE** should the policy layer specify any other trigger?  How do we control WHEN the policy becomes enforceable?  obviously the rule has to exist, but is there a downside to LOCKING a PDU or NCN right out of the gate?


