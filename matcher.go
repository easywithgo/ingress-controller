package ingress_controller

import (
	"regexp"
	"strings"
)

func (ing *IngressController) PrefixMatch(host, path string) (string, string, string, bool) {
	info := ing.router[host]
	for _, entry := range info.paths {
		if strings.HasPrefix(path, entry.Path) {
			return info.ingress.Namespace,
				info.ingress.Name,
				entry.Backend.ServiceName + "." + info.ingress.Namespace + ".svc:" + entry.Backend.ServicePort.String(),
				true
		}
	}
	return "", "", "", false
}


func (ing *IngressController) ExactMatch(host, path string) (string, string, string, bool) {
	info := ing.router[host]
	for _, entry := range info.paths {
		if path == entry.Path {
			return info.ingress.Namespace,
			info.ingress.Name,
			entry.Backend.ServiceName + "." + info.ingress.Namespace + ".svc:" + entry.Backend.ServicePort.String(),
			true
		}
	}
	return "", "", "", false
}

func (ing *IngressController) RegexMatch(host, path, pattern string) (string, string, string, bool) {
	info := ing.router[host]
	for _, entry := range info.paths {
		match, _ := regexp.MatchString(pattern, path)
		if match {
			return info.ingress.Namespace,
				info.ingress.Name,
				entry.Backend.ServiceName + "." + info.ingress.Namespace + ".svc:" + entry.Backend.ServicePort.String(),
				true
		}
	}
	return "", "", "", false
}

