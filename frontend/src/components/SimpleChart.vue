<script setup lang="ts">
import { computed } from 'vue'

interface ChartData {
  label: string
  value: number
  color?: string
}

interface Props {
  data: ChartData[]
  type: 'bar' | 'line' | 'pie'
  title?: string
  height?: number
}

const props = withDefaults(defineProps<Props>(), {
  height: 300,
  type: 'bar'
})

const chartHeight = computed(() => props.height || 300)

// 计算最大值，用于条形图缩放
const maxValue = computed(() => {
  return Math.max(...props.data.map(item => item.value))
})

// 为数据项生成颜色
const getColor = (index: number, customColor?: string) => {
  if (customColor) return customColor
  
  const colors = [
    '#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#9C27B0',
    '#FF9800', '#795548', '#607D8B', '#FF5722', '#8BC34A'
  ]
  return colors[index % colors.length]
}

// 计算百分比（用于饼图）
const totalValue = computed(() => {
  return props.data.reduce((sum, item) => sum + item.value, 0)
})

const getPercentage = (value: number) => {
  return Math.round((value / totalValue.value) * 100)
}

// 生成SVG路径（用于饼图）
const generatePieSlices = computed(() => {
  let startAngle = 0
  const slices: any[] = []
  
  for (let i = 0; i < props.data.length; i++) {
    const item = props.data[i]
    const percentage = getPercentage(item.value)
    const angle = (percentage / 100) * 360
    
    if (angle > 0) {
      slices.push({
        ...item,
        startAngle,
        endAngle: startAngle + angle,
        percentage,
        color: getColor(i, item.color)
      })
    }
    
    startAngle += angle
  }
  
  return slices
})

// 生成SVG路径字符串
const createArcPath = (startAngle: number, endAngle: number, radius: number = 80) => {
  const start = polarToCartesian(100, 100, radius, endAngle)
  const end = polarToCartesian(100, 100, radius, startAngle)
  const largeArcFlag = endAngle - startAngle <= 180 ? "0" : "1"
  
  return [
    "M", start.x, start.y,
    "A", radius, radius, 0, largeArcFlag, 0, end.x, end.y,
    "L", 100, 100,
    "Z"
  ].join(" ")
}

const polarToCartesian = (centerX: number, centerY: number, radius: number, angleInDegrees: number) => {
  const angleInRadians = (angleInDegrees - 90) * Math.PI / 180.0
  return {
    x: centerX + (radius * Math.cos(angleInRadians)),
    y: centerY + (radius * Math.sin(angleInRadians))
  }
}
</script>

<template>
  <div class="simple-chart">
    <h3 v-if="title" class="chart-title">{{ title }}</h3>
    
    <!-- 条形图 -->
    <div v-if="type === 'bar'" class="bar-chart" :style="{ height: (height || 300) + 'px' }">
      <div class="chart-container">
        <div class="y-axis">
          <div class="y-label">{{ maxValue }}</div>
          <div class="y-label">{{ Math.round(maxValue * 0.75) }}</div>
          <div class="y-label">{{ Math.round(maxValue * 0.5) }}</div>
          <div class="y-label">{{ Math.round(maxValue * 0.25) }}</div>
          <div class="y-label">0</div>
        </div>
        <div class="bars-container">
          <div v-for="(item, index) in data" :key="index" class="bar-item">
            <div 
              class="bar" 
              :style="{ 
                height: (item.value / maxValue) * (height - 80) + 'px',
                backgroundColor: getColor(index, item.color)
              }"
            >
              <span class="bar-value">{{ item.value }}</span>
            </div>
            <div class="bar-label">{{ item.label }}</div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 线形图 -->
    <div v-else-if="type === 'line'" class="line-chart" :style="{ height: height + 'px' }">
      <div class="chart-container">
        <div class="y-axis">
          <div class="y-label">{{ maxValue }}</div>
          <div class="y-label">{{ Math.round(maxValue * 0.75) }}</div>
          <div class="y-label">{{ Math.round(maxValue * 0.5) }}</div>
          <div class="y-label">{{ Math.round(maxValue * 0.25) }}</div>
          <div class="y-label">0</div>
        </div>
        <div class="line-container">
          <svg class="line-svg" :width="data.length * 80" :height="height - 60">
            <polyline
              :points="data.map((item, index) => `${index * 80 + 40},${height - 60 - (item.value / maxValue) * (height - 80)}`).join(' ')"
              stroke="#409EFF"
              stroke-width="2"
              fill="none"
            />
            <circle
              v-for="(item, index) in data"
              :key="index"
              :cx="index * 80 + 40"
              :cy="height - 60 - (item.value / maxValue) * (height - 80)"
              r="4"
              :fill="getColor(index, item.color)"
            />
          </svg>
          <div class="line-labels">
            <div 
              v-for="(item, index) in data" 
              :key="index" 
              class="line-label"
              :style="{ left: (index * 80 + 40) + 'px' }"
            >
              {{ item.label }}
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 饼图 -->
    <div v-else-if="type === 'pie'" class="pie-chart" :style="{ height: height + 'px' }">
      <div class="pie-container">
        <svg class="pie-svg" width="200" height="200" viewBox="0 0 200 200">
          <path
            v-for="(slice, index) in generatePieSlices"
            :key="index"
            :d="createArcPath(slice.startAngle, slice.endAngle)"
            :fill="slice.color"
            :stroke="'#fff'"
            stroke-width="2"
          />
        </svg>
        <div class="pie-legend">
          <div v-for="(item, index) in data" :key="index" class="legend-item">
            <div class="legend-color" :style="{ backgroundColor: getColor(index, item.color) }"></div>
            <span class="legend-label">{{ item.label }}</span>
            <span class="legend-value">{{ item.value }} ({{ getPercentage(item.value) }}%)</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.simple-chart {
  padding: 16px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chart-title {
  margin: 0 0 16px 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  text-align: center;
}

/* 条形图样式 */
.bar-chart {
  position: relative;
}

.chart-container {
  display: flex;
  height: 100%;
}

.y-axis {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 60px;
  padding: 20px 0 40px 0;
}

.y-label {
  font-size: 12px;
  color: #606266;
  text-align: right;
  padding-right: 8px;
}

.bars-container {
  display: flex;
  align-items: flex-end;
  flex: 1;
  padding: 20px 0 0 0;
  gap: 8px;
}

.bar-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.bar {
  position: relative;
  width: 100%;
  max-width: 40px;
  margin-bottom: 8px;
  border-radius: 4px 4px 0 0;
  transition: all 0.3s;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 4px;
}

.bar:hover {
  opacity: 0.8;
}

.bar-value {
  font-size: 12px;
  color: #fff;
  font-weight: 600;
}

.bar-label {
  font-size: 12px;
  color: #606266;
  text-align: center;
  margin-top: 8px;
  word-break: break-word;
}

/* 线形图样式 */
.line-chart {
  position: relative;
}

.line-container {
  position: relative;
  flex: 1;
  overflow-x: auto;
}

.line-svg {
  min-width: 100%;
}

.line-labels {
  position: relative;
  height: 40px;
  margin-top: 8px;
}

.line-label {
  position: absolute;
  bottom: 0;
  font-size: 12px;
  color: #606266;
  transform: translateX(-50%);
  white-space: nowrap;
}

/* 饼图样式 */
.pie-chart {
  display: flex;
  justify-content: center;
  align-items: center;
}

.pie-container {
  display: flex;
  align-items: center;
  gap: 32px;
}

.pie-svg {
  flex-shrink: 0;
}

.pie-legend {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 2px;
  flex-shrink: 0;
}

.legend-label {
  font-size: 14px;
  color: #303133;
  min-width: 80px;
}

.legend-value {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

@media (max-width: 768px) {
  .chart-container {
    flex-direction: column;
  }
  
  .y-axis {
    flex-direction: row;
    width: 100%;
    height: 30px;
    padding: 0;
  }
  
  .bars-container {
    padding: 10px 0 0 0;
  }
  
  .pie-container {
    flex-direction: column;
    gap: 16px;
  }
  
  .pie-svg {
    width: 150px;
    height: 150px;
  }
}
</style> 