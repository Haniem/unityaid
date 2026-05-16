export type CertificateType = 'hours' | 'participation'

export type Certificate = {
  id: string
  userId: string
  organizationId?: string | null
  type: CertificateType
  title: string
  description: string
  totalHours: number
  verifyCode: string
  issuedBy?: string | null
  issuedAt: string
  createdAt: string
}

export type CertificateGeneratePayload = {
  userId: string
  organizationId?: string | null
  type: CertificateType
  title?: string
  description?: string
}

export type CertificateListResponse = {
  items: Certificate[]
}
