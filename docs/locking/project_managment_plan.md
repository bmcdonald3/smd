# Locking Service Project Management

## High Level Tasks

### Development Tasks

 1. Create API specification (swagger)
 2. Create SQL model
 3. Cursory Exploration into technologies/packages that can help us
 4. Implement API -> Mk0
 5. Implement Agent -> Mk1? (we may defer implementation of this... so we would MANDATE that someone do this manually).
 6. Library -> so our go services can use it.
    7. other services have to integrate and change as well. 
 7. CLI -> This might be custom. Should use the Admin api? not the service api
 8. Investigate failure modes/ resiliency testing.  THIS thing HAS to work!  HIGHLY scalable; 
 9. RELATED:
   10. CAPMC change
   11. FAS change
   12. BOS change
   13. IFS change (technically just implement, b/c its so new
   14. SCSD change 


## Schedule
 2. Create SQL relational design - 2 engineers 2 days 
 1. Create API specification (swagger) - 2 engineers 2 days
 4. Implement API -> Mk0
 5. Implement Agent -> Mk1? (we may defer implementation of this... so we would MANDATE that someone do this manually).
 6. Implement Library -> so our go services can use it.
 7. CLI -> This might be custom. Should use the Admin api? not the service api.  If the CLI is stock it will take 1 engineer 0.5 days.  If it is custom it will take 1 engineer 5 days (per previous experience).
 8. Investigate failure modes/ resiliency testing.  THIS thing HAS to work!  HIGHLY scalable; 

## Impacts to Shasta v1.4

This document is being written 08/17/20;  Shasta v1.4 Freeze is 09/18/20.   Thats one month less 3 pto days.

Our thing has to be far enough along so other services can implement and test.

