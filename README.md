
# Jumbo petstore Test

## Running Locally
Make sure you have [Go](http://golang.org/doc/install) 

A .env file must be added to the root name ".env" , they is an .env-example file that can be used as a reference 

##Usage

"application/json" header required for requests 

## Notes

 * Uses GO
 * Uses AWS DynamoDB for the database
 * Live version was deployed on heroku 
 * Only the pets end point has been fully implemented 
 * Tested with postman
 
 I chose to use GO cause Iam currently leaning the language and was in the "GO mindset", also chose to use 
 DynamoDB because it would allow me to map the models , easily without creating to many tables and without need for an ORM,
 I think a Relational DB would have been more suitable though.
 
 
 ## Potential Issues / Required Fixes
 
 Pets->status should have a check that only allows a certain values eg only allow [ available, pending, sold ], is this a fixed enum value list ?
 
 A pet can only have one category ? 
 
 ApiResponse->message changed to an interface type , is sometime a string sometimes json 
 
 ApiResponse-> Code just returns  code 1 , Internal server errors (if they occur) might be broadcasting too much information to the user
 
 A lot of error handling code is repasting itself/breaking DRY principles, not sure how to handle this in GO yet
 
 Id conversion needs generic reusable function 
 
 Variable names need to be made more better ie err should be error .. but it seems like shorted variblke names are a thing in GO ??...
 
 Only returning JSON no XML being returned 
 
 On delete Item images are not being removed 
 
 Wanted to use S3 for file uploads but ran into the issue with the aws go sdk (its much easier with node js) 
 
 More Testing need but ran out of time 
 
 
 
