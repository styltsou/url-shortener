import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { useEffect, useState, useRef } from 'react'

interface ClicksGraphProps {
  data: Array<{ name: string; clicks: number }>
}

export function ClicksGraph({ data }: ClicksGraphProps) {
  const [colors, setColors] = useState({
    primary: 'hsl(var(--primary))',
    mutedForeground: 'hsl(var(--muted-foreground))',
    border: 'hsl(var(--border))',
    popover: 'hsl(var(--popover))',
    popoverForeground: 'hsl(var(--popover-foreground))',
  })

  const [isAnimationActive, setIsAnimationActive] = useState(false)
  const isInitialMount = useRef(true)
  const previousDataRef = useRef<string>('')

  useEffect(() => {
    // Get actual computed CSS variable values after mount
    const getColor = (varName: string) => {
      if (typeof window === 'undefined') return ''
      return getComputedStyle(document.documentElement).getPropertyValue(varName).trim()
    }

    setColors({
      primary: getColor('--primary') || 'hsl(var(--primary))',
      mutedForeground: getColor('--muted-foreground') || 'hsl(var(--muted-foreground))',
      border: getColor('--border') || 'hsl(var(--border))',
      popover: getColor('--popover') || 'hsl(var(--popover))',
      popoverForeground: getColor('--popover-foreground') || 'hsl(var(--popover-foreground))',
    })
  }, [])

  useEffect(() => {
    // Create a string representation of the data to detect changes
    const dataString = JSON.stringify(data)

    if (isInitialMount.current) {
      // On initial mount, disable animation and store the data
      isInitialMount.current = false
      previousDataRef.current = dataString
      setIsAnimationActive(false)
    } else if (previousDataRef.current !== dataString) {
      // Data has changed (due to filter), enable animation
      previousDataRef.current = dataString
      setIsAnimationActive(true)
    }
  }, [data])

  // Calculate Y-axis domain to ensure integer-only ticks
  const maxValue = Math.max(...data.map((d) => d.clicks), 0)
  const yAxisDomain: [number, number] = [0, Math.max(maxValue, 1)]

  return (
    <div className="h-[300px] w-full">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 10, right: 0, left: -20, bottom: 0 }}>
          <defs>
            <linearGradient id="colorClicks" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={colors.primary} stopOpacity={0.2} />
              <stop offset="95%" stopColor={colors.primary} stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" vertical={false} stroke={colors.border} />
          <XAxis
            dataKey="name"
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 12, fill: colors.mutedForeground }}
            dy={10}
          />
          <YAxis
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 12, fill: colors.mutedForeground }}
            allowDecimals={false}
            domain={yAxisDomain}
            tickFormatter={(value) => Math.floor(value).toString()}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: colors.popover,
              border: `1px solid ${colors.border}`,
              borderRadius: '8px',
              boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)',
              color: colors.popoverForeground,
            }}
            itemStyle={{ color: colors.popoverForeground }}
            labelStyle={{ color: colors.mutedForeground, marginBottom: '4px' }}
            cursor={{ stroke: colors.primary, strokeWidth: 1, strokeDasharray: '4 4' }}
            formatter={(value: number) => [Math.floor(value).toString(), 'Clicks']}
          />
          <Area
            type="monotone"
            dataKey="clicks"
            stroke={colors.primary}
            strokeWidth={3}
            fillOpacity={1}
            fill="url(#colorClicks)"
            isAnimationActive={isAnimationActive}
            animationDuration={200}
            connectNulls={false}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}

