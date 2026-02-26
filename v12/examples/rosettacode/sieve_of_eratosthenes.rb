def sieve(limit)
  return [] if limit < 2

  flags = Array.new(limit + 1, true)
  flags[0] = false
  flags[1] = false

  p = 2
  while p * p <= limit
    if flags[p]
      multiple = p * p
      while multiple <= limit
        flags[multiple] = false
        multiple += p
      end
    end
    p += 1
  end

  primes = []
  value = 2
  while value <= limit
    primes << value if flags[value]
    value += 1
  end

  primes
end

def show_primes(limit)
  primes = sieve(limit)
  puts "Primes up to #{limit}:"
  puts primes.inspect
  puts ""
end

show_primes(30)
show_primes(100)
show_primes(10000)
