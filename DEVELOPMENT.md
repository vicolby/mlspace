# Roadmap

## Projects
- [x] auth with keycloak
    - Debt: no data layer provided [x]
- [x] user logout
- [x] list of projects page
- [x] creating a project
    - Debt: validations inside handler and context vars inside middlware package [x][]
- [x] project page
    - [x] adding users to project
    - Debt: duplicates of project Id context setting inside handlers and services [x]
    - [x] deleting users from project
    - Debt: proper error and role access handling [x]
- [x] project deleting (only from projects list page for now)
- [x] create k8s namespace when project is created
  - debt: consider outbox pattern since there is a second api call after entity creation in DB
- project editing


## Disks
- [x] disk list
- [x] create disk modal
  - debt: error handling
- [x] create disk route
  - debt: ugly duplicates for fetching the project name by ID, and no server validation
- [x] delete disk route
