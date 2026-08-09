#!/data/data/com.termux/files/usr/bin/bash
# 江湖系列游戏启动器（2026-08-08 深夜交付）
cd "$(dirname "$0")"
echo ""
echo "  ╔══════════════════════════════╗"
echo "  ║   江 湖 系 列 · 游 戏 厅      ║"
echo "  ╚══════════════════════════════╝"
echo "  1) 江湖行       大型武侠开放世界 RPG"
echo "  2) 江湖·接暗器  快节奏反应小游戏"
echo "  3) 江湖·闭关    放置修炼摸鱼游戏"
echo "  4) 江湖行 --bot  AI 自动通关演示"
echo "  5) 退出"
echo ""
read -rp "  选择 [1-5]: " n
case "$n" in
  1) ./bin/jianghu ;;
  2) ./bin/dart ;;
  3) ./bin/closed ;;
  4) ./bin/jianghu --bot ;;
  *) exit 0 ;;
esac
