# Building a Cloud-Native Platform: From IaaS to PaaS

**Author:** Kevin Sinn

## Overview

**Meta-Track: Infrastructure → Platform → Product**

This track is about owning the full stack. You will get your hands on building a complete cloud platform stack end-to-end—starting from raw infrastructure, moving through Kubernetes, and ending with a production-grade PaaS product featuring APIs, UI, automation, and observability.

Every phase results in something tangible. By the end, you will have built and operated a real platform that reflects how modern cloud providers actually work.

Outcome: A fully functional PaaS on Kubernetes, running on OpenStack and SKE, complete with automation, observability, and developer tooling.

## Week 1

This week's Notion page contains all the knowledge, notes, and learnings obtained during Week 1 of the cloud-native platform journey.

📚 **[Week 1 Notion Page](https://www.notion.so/Notes-2edf7889a83b80d3bb24d64672a2613f?source=copy_link)**

### Goal

- [x] Understand and operate Infrastructure-as-a-Service by setting up and validating an OpenStack environment.

### Scope

- [x] **Understanding the OpenStack Architecture**
  - [x] How OpenStack is structured
  - [x] Responsibilities of core components:
    - [x] Virtual Machines
    - [x] Networks
    - [x] Storage and related services
- [x] **Installation and Basic Configuration** of a Single-node OpenStack cluster using DevStack
- [x] **Validation of the installation** by:
  - [x] Checking system logs
  - [x] Checking service status and health
- [x] **Deployment** of a virtual machine via OpenStack
- [x] **Architecture Analysis:**
  - [x] Analyze the current installation
  - [x] Create a technically correct architecture diagram of the setup

## Week 2

📚 **[Week 2 Notion Page](https://www.notion.so/Week-2-2f4f7889a83b800896e3e26754c9f499?source=copy_link)**

### Goal

- [x] Provision a Kubernetes cluster on OpenStack using Terraform - Infrastructure-as-Code (IaC)

### Scope

- [x] **Understanding Terraform** and Infrastructure as Code:
  - [x] What Infrastructure as Code is and what are its advantages
  - [x] Terraform basics:
    - [x] Providers
    - [x] Resources
    - [x] State
  - [x] Terraform workflow
- [x] **Provision a Virtual Machine** with Terraform using OpenStack APIs
- [x] **Installation and Configuration** of Kubernetes with a tool of your choice
- [x] **Core Kubernetes Concepts**
  - [x] Infrastructure Components: Control Plane and Worker Nodes
  - [x] Pods, Deployments, Services, Namespaces

### Bonus

- [x] **Fully Automated Installation:**
  - [x] Implement a single command to automatically provision the OpenStack Virtual Machine
  - [x] And the installation of the Kubernetes cluster

- [x] **Two-Node Kubernetes Cluster Setup:**
  - [x] Expand the installation script to provision two Virtual Machines
  - [x] And configure them to form a two-node Kubernetes cluster

## Week 3

📚 **[Week 3 Notion Page](https://www.notion.so/Week-3-2fbf7889a83b800196acd55504b150d9?source=copy_link)**

### Goal

- [x] Design and implement a Platform-as-a-Service offering on top of Kubernetes.

### Scope

- [x] **SKE Cluster Creation:** Using the STACKIT Terraform Provider to provision an SKE (STACKIT Kubernetes Engine) Cluster
- [x] **PaaS Product Implementation (e.g. Managed Database):** Design and technical implementation of a simple PaaS service.
  - [x] **Operator deployment:** Provisioning of an Operator
  - [x] **Product Component Management:** Utilization of Custom Kubernetes Resources (CRs) for the provisioning and management of product components
  - [x] **Connectivity:** Documentation and demonstration of connecting to and using the PaaS product
- [x] **Understanding Kubernetes Concepts:** Deepening knowledge of Custom Resource Definitions (CRDs) and the functioning of Operators (Reconciler Pattern)

### Bonus

- [x] **Automating the Deployment:** Introduction of a GitOps approach and CI/CD integration for automated provisioning of the SKE and the PaaS service

## Week 4

📚 **[Week 4 Notion Page](https://www.notion.so/Week-4-2fef7889a83b80b9b7bfd42b9cfbdfc8?source=copy_link)**

### Goal

- [ ] Expose the PaaS product through a clean, production-ready RESTful API to enable automated provisioning and seamless integration.

### Scope

- [ ] **API Development for Product Instances**: Development of a RESTful API that provides the following functions: creation, deletion, and listing of PaaS product instances, as well as retrieval of the associated connection and access data.
    - [ ] API Specification: The functionality of the API must be documented in the form of OpenAPI specifications
- [ ] **Unit Tests:** Implementation of simple Unit Tests for each endpoint of the developed API
- [ ] **Docker Container Image** for the API creation, upload to the STACKIT Container Registry, and provisioning of the API via SKE
- [ ] **Understanding the Creation Process (Create Flow):**
    - [ ] * **Flowchart for the creation of a product instance:** Visualization of the individual steps
    - [ ] * **Basic understanding of how a RESTful API works**

### Bonus

- [ ] **Automated API Deployment**: Integration of the deployment of the RESTful API into the existing GitOps infrastructure
- [ ] **Auto-Scaling and Performance Tests**:
    - [ ] * **Horizontal Pod Autoscaler (HPA):** Configuration of the Kubernetes HPA for automatic scaling of the Control Plane RESTful API
    - [ ] * **Performance Tests**: Implementation of performance tests for the RESTful API to verify the functionality of the HPA
- [ ] **Update Functionality**: Implementation and testing of an API endpoint that allows the updating of access data and details of the product instance


## Week 5

_To be updated..._

## Week 6

_To be updated..._

