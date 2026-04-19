#!/usr/bin/env bash
# ============================================================
#  Lybel Skills - macOS/Linux Installer
#  Baixa e instala a skill lybel-docs no Claude Desktop/Code
# ============================================================

set -euo pipefail

# Cores (com fallback se o terminal não suportar)
if [[ -t 1 ]] && command -v tput >/dev/null 2>&1; then
  BOLD=$(tput bold)
  RESET=$(tput sgr0)
  GREEN=$(tput setaf 2)
  CYAN=$(tput setaf 6)
  YELLOW=$(tput setaf 3)
  RED=$(tput setaf 1)
  GRAY=$(tput setaf 8 2>/dev/null || echo "")
else
  BOLD=""; RESET=""; GREEN=""; CYAN=""; YELLOW=""; RED=""; GRAY=""
fi

# Detecta OS+arch para baixar o ZIP com binário correto
_detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  case "${os}" in
    darwin) os="darwin" ;;
    linux)  os="linux" ;;
    *)      echo "unknown-unknown"; return ;;
  esac
  case "${arch}" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *)             arch="amd64" ;;  # fallback
  esac
  echo "${os}-${arch}"
}
PLATFORM="$(_detect_platform)"
RELEASE_URL="https://github.com/lybel-app/skills/releases/latest/download/lybel-docs-${PLATFORM}.zip"
INSTALL_DIR="${HOME}/.claude/skills/lybel-docs"
TMP_DIR="$(mktemp -d -t lybel-docs-XXXXXX)"
TMP_ZIP="${TMP_DIR}/lybel-docs.zip"
TMP_EXTRACT="${TMP_DIR}/extract"

cleanup() {
  rm -rf "${TMP_DIR}" 2>/dev/null || true
}
trap cleanup EXIT

err() {
  echo ""
  echo "${RED}============================================================${RESET}"
  echo "${RED}  ERRO na instalação${RESET}"
  echo "${RED}============================================================${RESET}"
  echo ""
  echo "${RED}${1}${RESET}"
  echo ""
  echo "${YELLOW}Dicas:${RESET}"
  echo "  - Verifique sua conexão com a internet."
  echo "  - Se a primeira release ainda não foi publicada, aguarde ou contate o Diego."
  echo "  - Confirme que você tem curl ou wget instalado."
  echo ""
  exit 1
}

echo ""
echo "${BOLD}============================================================${RESET}"
echo "${BOLD}  Instalador Lybel Skills (lybel-docs)${RESET}"
echo "${BOLD}============================================================${RESET}"
echo ""

# Detecta OS
OS="$(uname -s)"
case "${OS}" in
  Darwin*) OS_LABEL="macOS" ;;
  Linux*)  OS_LABEL="Linux" ;;
  *)       err "Sistema operacional não suportado: ${OS}. Use install.bat no Windows." ;;
esac
echo "${CYAN}[info] Sistema detectado:${RESET} ${OS_LABEL}"

# Detecta downloader
if command -v curl >/dev/null 2>&1; then
  DOWNLOADER="curl"
elif command -v wget >/dev/null 2>&1; then
  DOWNLOADER="wget"
else
  err "Nenhum downloader encontrado. Instale curl ou wget e tente de novo."
fi

# Detecta unzip
if ! command -v unzip >/dev/null 2>&1; then
  err "'unzip' não encontrado. Instale com: brew install unzip (mac) ou apt-get install unzip (linux)."
fi

# 1. Preparar diretório
echo "${CYAN}[1/4] Preparando diretório de instalação...${RESET}"
echo "${GRAY}      -> ${INSTALL_DIR}${RESET}"
rm -rf "${INSTALL_DIR}"
mkdir -p "${INSTALL_DIR}"

# 2. Download
echo "${CYAN}[2/4] Baixando última versão...${RESET}"
echo "${GRAY}      -> ${RELEASE_URL}${RESET}"
if [[ "${DOWNLOADER}" == "curl" ]]; then
  curl -fsSL "${RELEASE_URL}" -o "${TMP_ZIP}" || err "Download falhou. URL: ${RELEASE_URL}"
else
  wget -q "${RELEASE_URL}" -O "${TMP_ZIP}" || err "Download falhou. URL: ${RELEASE_URL}"
fi

# 3. Extrair
echo "${CYAN}[3/4] Extraindo arquivos...${RESET}"
mkdir -p "${TMP_EXTRACT}"
unzip -q "${TMP_ZIP}" -d "${TMP_EXTRACT}" || err "Falha ao extrair o ZIP."
cp -R "${TMP_EXTRACT}/." "${INSTALL_DIR}/"

# Torna binário executável se existir
if [[ -f "${INSTALL_DIR}/bin/lybel-docs" ]]; then
  chmod +x "${INSTALL_DIR}/bin/lybel-docs"
fi

# 4. Done
echo "${CYAN}[4/4] Finalizando...${RESET}"
echo ""
echo "${GREEN}============================================================${RESET}"
echo "${GREEN}  Instalação concluída com sucesso!${RESET}"
echo "${GREEN}============================================================${RESET}"
echo ""
echo "${YELLOW}Próximos passos:${RESET}"
echo "  1. Reinicie o Claude Desktop/Code (feche e abra de novo)."
echo "  2. Se estiver usando Claude Code, rode:  ${BOLD}claude${RESET}  em um terminal novo."
echo "  3. Configure a integração Atlassian (OAuth) nas Settings do Claude."
echo "  4. Pergunte em linguagem natural, ex:"
echo "     ${GRAY}\"onde cadastro um novo advogado?\"${RESET}"
echo "     ${GRAY}\"me dá a página de parceiros\"${RESET}"
echo ""
