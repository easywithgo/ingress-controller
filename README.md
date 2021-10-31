# ingress-controller
monitor changes to an ingress in kubernetes

This is meant to automatically monitor the changes in kubernetes ingress definitions. 
This will manage an inmemory map of all the ingresses currently available within a kube cluster. 

# Matcher
Three matchers are available PrefixMatch, ExactMatch and RegexMatch are available. 

Read Issues to identify what needs to be done.
