interface Finding {
  description: string
  risk_level: 'High' | 'Medium' | 'Low'
  remediation: string
}

interface FindingsReportProps {
  findings: Finding[]
}

const riskConfig = {
  High: {
    label: 'High',
    indicatorClass: 'bg-red-500',
    badgeClass: 'bg-red-100 text-red-700 border-red-200',
    borderClass: 'border-red-200',
  },
  Medium: {
    label: 'Medium',
    indicatorClass: 'bg-yellow-400',
    badgeClass: 'bg-yellow-100 text-yellow-700 border-yellow-200',
    borderClass: 'border-yellow-200',
  },
  Low: {
    label: 'Low',
    indicatorClass: 'bg-green-500',
    badgeClass: 'bg-green-100 text-green-700 border-green-200',
    borderClass: 'border-green-200',
  },
}

export default function FindingsReport({ findings }: FindingsReportProps) {
  if (findings.length === 0) {
    return (
      <div className="flex items-center gap-2 p-4 bg-green-50 border border-green-200 rounded-xl text-green-700 text-sm">
        <span className="w-2 h-2 rounded-full bg-green-500 inline-block" />
        No security issues found. This policy looks good!
      </div>
    )
  }

  return (
    <div className="space-y-3">
      <p className="text-sm text-gray-500">{findings.length} finding{findings.length !== 1 ? 's' : ''} detected</p>
      {findings.map((finding, i) => {
        const cfg = riskConfig[finding.risk_level] ?? riskConfig.Low
        return (
          <div
            key={i}
            className={`border ${cfg.borderClass} rounded-xl p-4`}
          >
            <div className="flex items-start gap-3">
              <span
                className={`mt-1 w-2.5 h-2.5 rounded-full flex-shrink-0 ${cfg.indicatorClass}`}
                aria-label={`${finding.risk_level} risk`}
              />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className={`text-xs font-semibold px-2 py-0.5 rounded border ${cfg.badgeClass}`}>
                    {cfg.label}
                  </span>
                </div>
                <p className="text-sm text-gray-800 font-medium mb-2">{finding.description}</p>
                <div className="bg-gray-50 rounded-lg p-3">
                  <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">Remediation</p>
                  <p className="text-sm text-gray-700">{finding.remediation}</p>
                </div>
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}
