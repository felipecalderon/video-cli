# 📊 FODA (Análisis estratégico)

## 🟢 Fortalezas

- Enfoque innovador (percepción vs render tradicional)
- Diferenciación clara frente a proyectos existentes
- Escalable (de experimento a librería seria)
- Buen fit con lenguajes de alto rendimiento (Go, C++)

---

## 🟡 Oportunidades

- Crear un “nuevo estándar” de render en terminal
- Uso en:
  - herramientas CLI visuales
  - arte generativo
  - demos técnicas virales

- Posible open source atractivo
- Integración con streaming en terminal

---

## 🔴 Debilidades

- Dependencia fuerte de:
  - rendimiento de terminal
  - implementación del emulador

- Difícil de testear objetivamente (percepción ≠ métricas claras)
- Complejidad alta en tuning de parámetros visuales
- Resultados pueden variar entre usuarios

---

## ⚫ Amenazas

- Limitaciones físicas del terminal (hard cap)
- Fatiga visual si el dithering es agresivo
- Comparación con soluciones gráficas reales (siempre perderás)
- Abandono del proyecto por scope creep

---

# ⚠️ Riesgos técnicos clave

### 1. “Demasiado inteligente para su propio bien”

- Overengineering temprano
- Solución compleja antes de validar lo básico

👉 Mitigación: MVP simple primero

---

### 2. Ruido visual

- Dithering mal calibrado → imagen molesta

👉 Mitigación:

- presets
- testing visual constante

---

### 3. Cuello de botella en output

- stdout limita FPS

👉 Mitigación:

- diff rendering obligatorio

---

# 🧠 Principios de diseño (muy importante)

1. **El ojo es el target, no el pixel**
2. **Menos datos ≠ peor calidad percibida**
3. **La ilusión importa más que la precisión**
4. **La estabilidad visual > detalle máximo**
5. **Cada terminal es un entorno distinto**

---
