# K8s operator notes

- Operators:
  - Controller: Kube Controller Manager, 
  - Forever running loop:
    - Watches for changes to resources
    - Compares the desired state (spec) with the current state (status)
    - Takes action to reconcile the two states (we write our business logic here)

## Properties of operators

- Idempotent
  - Happy path: If there are no changes, do nothing and exits gracefully.
  - Change path: If there are changes, make the necessary updates to reach the desired state. If update works, update the .status. If update fails, log the error and retry later.
  - <img width="1370" height="897" alt="image" src="https://github.com/user-attachments/assets/4284094e-db2a-4c6f-8e3b-be0fbef61ea4" />

* Operator is CR + CRD
* Controller is this how part

* Think of k8s as SDK
