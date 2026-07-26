/** internal/handler/auth_handler.go + internal/service/auth_service.go */

export interface LoginPayload {
  username: string
  password: string
}

export interface RegisterPayload {
  username: string
  password: string
  nickname: string
}

export interface RegisterOutput {
  user_id: number
  player_id: number
}

export interface LoginOutput {
  token: string
  expires_in: number
  player: {
    player_id: number
    nickname: string
    level: number
  }
}
