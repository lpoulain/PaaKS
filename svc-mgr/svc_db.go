package main

import (
	"database/sql"
	"fmt"
)

func queryServicesInDatabase() (map[string]Service, error) {
	connStr := getConnectionString()
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("ERROR connecting to the database:", err.Error())
		return nil, err
	}

	sql := "SELECT name, tenant FROM admin.services"
	rows, err := conn.Query(sql)

	if err != nil {
		fmt.Println("ERROR querying data:", err.Error())
		return nil, err
	}

	services := make(map[string]Service)

	defer conn.Close()
	defer rows.Close()
	for rows.Next() {
		var name string
		var tenant string
		if err := rows.Scan(&name, &tenant); err != nil {
			return map[string]Service{}, err
		}

		key := fmt.Sprintf("tnt-%s-%s", tenant[:8], name)
		services[key] = Service{
			name:   name,
			tenant: tenant,
		}
	}

	return services, nil
}

func createServiceInDatabase(name string, tenant string) error {
	connStr := getConnectionString()
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("ERROR connecting to the database:", err.Error())
		return err
	}

	sql := "INSERT INTO admin.services(name, tenant) VALUES($1, $2)"
	stmt, err := conn.Prepare(sql)
	if err != nil {
		return nil
	}

	_, err = stmt.Query(name, tenant)

	return err
}

func deleteServiceInDatabase(name string, tenant string) error {
	connStr := getConnectionString()
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("ERROR connecting to the database:", err.Error())
		return err
	}

	sql := "DELETE FROM admin.services WHERE name=$1 AND tenant=$2"
	stmt, err := conn.Prepare(sql)
	if err != nil {
		return nil
	}

	_, err = stmt.Query(name, tenant)

	return err
}
