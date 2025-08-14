package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		printHelp()
		return
	}

	fmt.Println("🔐 Gerador de Chaves Seguras para SIGA Core Gateway")
	fmt.Println("==================================================")

	// Gerar chave de 32 bytes (256 bits) para tokens
	tokenKey, err := generateSecureKey(32)
	if err != nil {
		log.Fatalf("Erro ao gerar chave: %v", err)
	}

	fmt.Printf("\n✅ Chave Simétrica Gerada (32 caracteres):\n")
	fmt.Printf("TOKEN_SYMMETRIC_KEY=%s\n", tokenKey)

	fmt.Printf("\n📋 Para usar em produção:\n")
	fmt.Printf("export TOKEN_SYMMETRIC_KEY=\"%s\"\n", tokenKey)

	fmt.Printf("\n⚠️  IMPORTANTE:\n")
	fmt.Printf("- Mantenha esta chave SECRETA\n")
	fmt.Printf("- NÃO armazene em repositórios Git\n")
	fmt.Printf("- Use um gerenciador de segredos em produção\n")
	fmt.Printf("- Considere rotacionar a chave periodicamente\n")

	// Opcional: gerar múltiplas chaves para diferentes ambientes
	if len(os.Args) > 1 && os.Args[1] == "--multiple" {
		fmt.Printf("\n🔄 Chaves adicionais para rotação:\n")
		for i := 1; i <= 3; i++ {
			key, err := generateSecureKey(32)
			if err != nil {
				log.Printf("Erro ao gerar chave %d: %v", i, err)
				continue
			}
			fmt.Printf("BACKUP_KEY_%d=%s\n", i, key)
		}
	}
}

// generateSecureKey gera uma chave criptograficamente segura
func generateSecureKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("falha ao gerar bytes aleatórios: %w", err)
	}

	// Converter para string hexadecimal legível
	return hex.EncodeToString(bytes)[:length], nil
}

func printHelp() {
	fmt.Println("Gerador de Chaves Seguras - SIGA Core Gateway")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run scripts/generate-key.go           # Gera uma chave")
	fmt.Println("  go run scripts/generate-key.go --multiple # Gera múltiplas chaves")
	fmt.Println("  go run scripts/generate-key.go --help     # Mostra esta ajuda")
	fmt.Println()
	fmt.Println("EXEMPLOS:")
	fmt.Println("  # Gerar e usar diretamente:")
	fmt.Println("  export TOKEN_SYMMETRIC_KEY=$(go run scripts/generate-key.go | grep TOKEN_SYMMETRIC_KEY | cut -d'=' -f2)")
	fmt.Println()
	fmt.Println("  # Compilar e usar:")
	fmt.Println("  go build -o bin/generate-key scripts/generate-key.go")
	fmt.Println("  ./bin/generate-key")
}
