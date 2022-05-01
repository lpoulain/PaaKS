package main

import (
	"database/sql"
	"fmt"
)

func serviceConstructor(rows *sql.Rows) (Service, error) {
	var name string
	var tenant string
	if err := rows.Scan(&name, &tenant); err != nil {
		return Service{}, err
	}

	return Service{
		name:   name,
		tenant: tenant,
	}, nil
}

func queryServicesInDatabase() (map[string]Service, error) {
	services, err := query[Service]("SELECT name, tenant FROM admin.services", serviceConstructor)
	if err != nil {
		return nil, err
	}

	serviceMap := make(map[string]Service)

	for _, service := range services {
		key := fmt.Sprintf("tnt-%s-%s", service.tenant[:8], service.name)
		serviceMap[key] = service
	}

	return serviceMap, nil
}

func createServiceInDatabase(name string, tenant string) error {
	return exec("INSERT INTO admin.services (name, tenant) VALUES($1, $2)", name, tenant)
}

func deleteServiceInDatabase(name string, tenant string) error {
	return exec("DELETE FROM admin.services (name, tenant) VALUES($1, $2)", name, tenant)
}
