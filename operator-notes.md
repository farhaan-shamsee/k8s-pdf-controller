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
  - Infinite Loop Path(Sad Path): Only write to API server when there are changes. If no changes, do nothing. This prevents infinite loops of updates.
    - <img width="1804" height="685" alt="image" src="https://github.com/user-attachments/assets/1747b6d3-99c1-472b-aee7-9dde6fe90041" />
 

* Operator is CR + CRD
* Controller is this how part

* Think of k8s as SDK

## Operator Functions

1. "k8s.io/client-go/plugin/pkg/client/auth": Imports all client auth plugins (e.g., Azure, GCP, OIDC, etc.)
2. `cmd/main.go`: Entry point of the application. Sets up the manager and starts the controller.
3. `config`: Contains configuration files for the operator.