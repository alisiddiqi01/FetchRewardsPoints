# Fetch Rewards Points 
I have implemented a service to manage point transactions for users and their payers as outlined  
in the description. The service uses in-memory storage to keep track of a single user's account  
information, but extending to multiple users is also possible with some modifications to the storage  
strategy.

## Storage Strategy
The information stored by the service includes the point payers (i.e., a company like Dannon), their  
total point balances, and a history of the user's transactions.  
### Payers and Balances
Payers and their total point balances (how many points the user has from the payer, i.e., a user  
having 500 points from Unilever and 450 from Fairlife) are stored in a string-int map, where each  
string represents a payer's name ("UNILEVER", "FAIRLIFE") and the mapped int value for a payer  
represents the payer's total balance for the user, combining all current transaction history for the  
user.  
### Transactions
Transactions are stored in Transaction objects, containing values for the payer (string),  
point value (int), and timestamp (string) of the transaction. These objects are kept in a sorted array,  
which is sorted after calls to add Transactions are made. While it would be more efficient to use a  
structure like a min-heap because of the spending constraint regarding the order of transactions  
(oldest first) and the faster average insert time for min-heaps, I am not very experienced in Go  
currently and I felt a sorted array, while less efficient, would serve its purpose for this small project.  

Beyond the regular Transaction objects, we also store Balance objects, which contain a payer  
(string), and point value (int) corresponding to the amount a payer was deducted when a user  
spends points.

## Adding Points

To add points a `POST` request must be routed to the route `/add` with a json object  
containing the fields`"payer"`, `"points"`, and `"timestamp"`. The service uses these fields to  
populate a Transaction object and add it to the sorted array of Transactions, as well as updating  
the map of balances.


## Spending Points

Points are spent by sending a `POST` request to the route `/spend` with a json object containing  
a `"points"` field. Using this amount of points to spend, the service calls the corresponding  
spend functions.

The spending strategy proceeds as follows:  

* We first iterate through our sorted Transactions array in reverse (newest to oldest), populating  
an availableHere array with the amount of points available at each timestamp. This is important  
because it allows us to avoid any negative balances that may result from future deductions. For  
example, while a user may have +300 points during a given transaction, a future transaction  
of -250 points means that they only have 50 points available (300-250) at the timestamp of  
the +300 point transaction. If we allow them to spend all 300 instead, at the timestamp of  
the -250 point transaction the balance of the user for that payer is now negative.  

* After populating our availableHere array, we iterate through our sorted Transactions array  
(oldest to newest), checking the amount of points that are available at each point and  
adjusting the remaining balance accordingly. We also maintain a list of the deductions that  
may occur, including the payers who are contributing to the spending and their respective  
point contributions.

* If the remaining balance hits 0, we create a list of Balance objects corresponding to the  
deductions we recorded for each involved payer, as well as Transaction objects to update our  
transaction history. We add the Transaction objects, update our map of `"payer":point` pairs,  
and return the Balance list.

* If the remaining balance doesn't hit 0, there were not enough points to spend and we return  
an error.

## Getting the Point Balance

The point balance for each payer can be retrieved by making a `GET` request to the route `/points`  
The service returns the map of `"payer":points` corresponding to the updated point balances  
of each payer.


## Running and Testing the Service

First ensure that your system has Docker and Docker Compose installed and running. Then,  
navigate to the project directory and run the command `docker-compose up -d`  

By default, the service runs and then can be accessed through the previously mentioned routes  
at http://localhost:8080.  

To run the tests, edit the `start.sh` file in the project directory to uncomment the test command  
`go test` and comment out the run command `./FetchRewardsPoints`. Run  
`docker-compose build` in the project directory and again run the `docker-compose up` command  
to see the test results. 

To clean up the environment, simply run `docker-compose down` in the project directory.

## Improvements
Possible improvements to this service include firstly modifying the storage structure to account  
for multiple users and efficiently inserting transactions for each. The current sorted array  
approach works, but would be costly when attempting to scale as it requires additional time to  
both resize and sort itself as compared to a min-heap. 

Additionally, once the frequency of add and spend operations is established for an average user,  
we can make further determinations about whether it might be beneficial to augment the spend  
functions with additional data structures that store information about the amount of points  
available at different points in time. This could provide performance advantages as it is calculated  
upon every spend operation and would be beneficial if spend operations are frequent.
