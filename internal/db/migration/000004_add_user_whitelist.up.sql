-- Adiciona coluna para whitelist de usuários
ALTER TABLE "users" ADD COLUMN "is_whitelisted" boolean NOT NULL DEFAULT false;

-- Cria índice para consultas rápidas de usuários whitelistados
CREATE INDEX "idx_users_whitelisted" ON "users" ("is_whitelisted") WHERE "is_whitelisted" = true; 