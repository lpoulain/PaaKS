package paaks

import "os"

func getSecretKey() string {
    return os.Getenv("SECRET_KEY")
}
