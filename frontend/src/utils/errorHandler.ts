import type { AxiosError } from 'axios'
import { ElMessage } from 'element-plus'

// 错误处理工具类
export class ErrorHandler {
  /**
   * 从错误对象中提取错误消息
   */
  static extractErrorMessage(error: any): string {
    // 如果是AxiosError
    if (error.response?.data) {
      // 尝试不同的错误消息字段
      return error.response.data.message || 
             error.response.data.error || 
             error.response.data.msg ||
             error.response.data.detail ||
             error.response.data.details ||
             `请求失败 (${error.response.status})`
    }
    
    // 如果是普通Error对象
    if (error.message) {
      return error.message
    }
    
    // 如果是字符串
    if (typeof error === 'string') {
      return error
    }
    
    // 默认错误消息
    return '操作失败'
  }

  /**
   * 显示错误消息
   */
  static showError(error: any, defaultMessage?: string): void {
    const message = defaultMessage || this.extractErrorMessage(error)
    ElMessage.error(message)
  }

  /**
   * 显示成功消息
   */
  static showSuccess(message: string): void {
    ElMessage.success(message)
  }

  /**
   * 显示警告消息
   */
  static showWarning(message: string): void {
    ElMessage.warning(message)
  }

  /**
   * 处理API错误的通用方法
   */
  static handleApiError(error: any, context?: string): void {
    console.error(`${context || 'API'} 错误:`, error)
    
    // 如果是404错误，可能需要特殊处理
    if (error.response?.status === 404) {
      const message = this.extractErrorMessage(error)
      ElMessage.error(message)
      return
    }
    
    // 其他错误直接显示
    this.showError(error)
  }

  /**
   * 处理表单验证错误
   */
  static handleValidationError(error: any): Record<string, string[]> {
    const errors: Record<string, string[]> = {}
    
    if (error.response?.data?.errors) {
      // 如果后端返回字段级别的错误
      return error.response.data.errors
    }
    
    if (error.response?.data?.details && Array.isArray(error.response.data.details)) {
      // 处理详细错误列表
      error.response.data.details.forEach((detail: any) => {
        if (detail.field && detail.message) {
          if (!errors[detail.field]) {
            errors[detail.field] = []
          }
          errors[detail.field].push(detail.message)
        }
      })
    }
    
    return errors
  }

  /**
   * 根据HTTP状态码获取默认错误消息
   */
  static getDefaultErrorByStatus(status: number): string {
    switch (status) {
      case 400:
        return '请求参数错误'
      case 401:
        return '未授权访问'
      case 403:
        return '权限不足'
      case 404:
        return '资源不存在'
      case 409:
        return '资源冲突'
      case 422:
        return '请求数据验证失败'
      case 429:
        return '请求过于频繁'
      case 500:
        return '服务器内部错误'
      case 502:
        return '网关错误'
      case 503:
        return '服务暂不可用'
      case 504:
        return '网关超时'
      default:
        return `请求失败 (${status})`
    }
  }
}

// 导出便捷方法
export const showError = (error: any, defaultMessage?: string) => 
  ErrorHandler.showError(error, defaultMessage)

export const showSuccess = (message: string) => 
  ErrorHandler.showSuccess(message)

export const showWarning = (message: string) => 
  ErrorHandler.showWarning(message)

export const handleApiError = (error: any, context?: string) => 
  ErrorHandler.handleApiError(error, context)

export const extractErrorMessage = (error: any) => 
  ErrorHandler.extractErrorMessage(error)
