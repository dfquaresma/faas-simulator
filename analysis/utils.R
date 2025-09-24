library(dplyr)
library(ggplot2)
library(lubridate)
library(tidyr)

preprocess_data <- function(df) {
  rm_smallsample_functions <- function(df) {
    # Return the sub_df if sub_df has more than 1000 rows, otherwise NULL
    if (nrow(df) > 40000) {
      return(df)
    } else {
      return(NULL)
    }
    return(df)
  }
  removezeros_countrows <- function(df) {
    df = df[df$duration != 0, ]
    df$rows <- nrow(df)
    return(df)
  }
  calculate_percentiles <- function(df) {
    p50    <- quantile(df$duration, probs = 0.50)
    p95    <- quantile(df$duration, probs = 0.95)
    p99    <- quantile(df$duration, probs = 0.99)
    p999    <- quantile(df$duration, probs = 0.999)
    p9999    <- quantile(df$duration, probs = 0.9999)
    p100   <- quantile(df$duration, probs = 1)
    df$start_timestamp = df$end_timestamp - df$duration
    df$p50   <- p50
    df$p95   <- p95
    df$p99   <- p99
    df$p999  <- p999
    df$p9999 <- p9999
    df$p100  <- p100
    return(df)
  }
  
  # Split the df dataframe by values in func and apply filter_functions to each sub-dataframe
  inv2021_split <- lapply(split(df, df$func), rm_smallsample_functions)
  # Remove null entries in inv2021_split
  inv2021_split_nonull <- inv2021_split[!sapply(inv2021_split, is.null)]
  inv2021_split_nozeros <- lapply(inv2021_split_nonull, removezeros_countrows)
  inv2021_split_processed <- lapply(inv2021_split_nozeros, calculate_percentiles)
  
  # Merge data
  inv2021_merged = do.call(rbind, inv2021_split_processed)
  inv2021_merged = data.frame(
    app = inv2021_merged$app,
    func = inv2021_merged$func,
    rows = inv2021_merged$rows,
    start_timestamp = inv2021_merged$start_timestamp,
    duration = inv2021_merged$duration,
    end_timestamp = inv2021_merged$end_timestamp,
    
    p50 = inv2021_merged$p50,
    p95 = inv2021_merged$p95,
    p99 = inv2021_merged$p99,
    p999 = inv2021_merged$p999,
    p9999 = inv2021_merged$p9999,
    p100 = inv2021_merged$p100
  )
  
  # sort by start_timestamp
  inv2021_merged = select(
    inv2021_merged[order(inv2021_merged$start_timestamp),], 
    app, func, rows,
    start_timestamp, duration, end_timestamp,
    p50, p95, p99, p999, p9999, p100
  )
  return(inv2021_merged)
}

plot_requests_workload <- function(df, bin_size = "10 min", title = "Requests Over Time") {
  # converter start_timestamp em POSIXct (ajuste origin conforme necessário)
  df$time_readable <- as.POSIXct(df$start_timestamp, origin = "1970-01-01", tz = "UTC")
  
  binned_df <- df %>%
    mutate(bin_time = floor_date(time_readable, unit = bin_size)) %>%  # arredonda para o início do bin
    group_by(bin_time) %>%
    summarise(requests = n(), .groups = "drop")
  
  ggplot(binned_df, aes(x = bin_time, y = requests)) +
    geom_line(color = "steelblue", size = 1) +
    geom_point(color = "steelblue", size = 1) +
    scale_x_datetime(
      date_labels = "%d %b\n%H:%M",
      date_breaks = "2 days"
    ) +
    labs(
      title = title,
      x = "Time",
      y = "Amount of Requests"
    ) +
    theme_minimal(base_size = 14) +
    theme(
      plot.title = element_text(hjust = 0.5, face = "bold"),
      legend.position = "none"
    )
}

plot_histogram <- function(df1, df2, df1_name, df2_name, title, colors) {
  p50_df1 = quantile(df1, probs = 0.5, na.rm = TRUE)
  p50_df2 = quantile(df2, probs = 0.5, na.rm = TRUE)
  p95_df1 = quantile(df1, probs = 0.95, na.rm = TRUE)
  p95_df2  = quantile(df2, probs = 0.95, na.rm = TRUE)
  p999_df1 = quantile(df1, probs = 0.999, na.rm = TRUE)
  p999_df2  = quantile(df2, probs = 0.999, na.rm = TRUE)
  p100_df1 = quantile(df1, probs = 1.0, na.rm = TRUE)
  p100_df2  = quantile(df2, probs = 1.0, na.rm = TRUE)
  
  df1_name_upto95 = paste(df1_name, " up to p95", sep="")
  df2_name_upto95 = paste(df2_name, " up to p95", sep="")

  # Prepare data with 4 different traces
  data_combined <- data.frame(
    duration = c(df1, 
                 df1[df1 <= p95_df1], 
                 df2,
                 df2[df2 <= p95_df2]),
    type = factor(c(rep(df1_name, length(df1)),
                    rep(df1_name_upto95, length(df1[df1 <= p95_df1])),
                    rep(df2_name, length(df2)),
                    rep(df2_name_upto95, length(df2[df2 <= p95_df2]))),
                  levels = c(df1_name, df1_name_upto95,
                             df2_name, df2_name_upto95)
                  )
  )
  
  
  max_y <- data_combined %>% group_by(type) %>% summarise(ymax = max(hist(duration, breaks = 30, plot = FALSE)$counts))
  
  # Create color mapping for facets
  colors_mapping <- c(
    df1_name        = colors[1],
    df1_name_upto95 = colors[3],
    df2_name        = colors[2],
    df2_name_upto95 = colors[4]
  )
  names(colors_mapping) <- c(df1_name, df1_name_upto95, df2_name, df2_name_upto95)
  
  # Create the plot
  ggplot(data_combined, aes(x = duration, fill = type)) +
    geom_histogram(alpha = 0.8, bins = 30) +
    facet_wrap(~type, scales = "free", nrow = 2, ncol = 2) +
    scale_fill_manual(values = colors_mapping) +
    
    
    # Add p50 lines only for truncated traces
    geom_vline(data = data.frame(type = df1_name_upto95, p50 = p50_df1), 
               aes(xintercept = p50), color = colors[5], linetype = "dashed", size = 1.2) +
    geom_vline(data = data.frame(type = df2_name_upto95, p50 = p50_df2), 
               aes(xintercept = p50), color = colors[6], linetype = "dashed", size = 1.2) +
    geom_vline(data = data.frame(type = df1_name_upto95, p95 = p95_df1), 
               aes(xintercept = p95), color = colors[5], linetype = "dotted", size = 1.2) +
    geom_vline(data = data.frame(type = df2_name_upto95, p95 = p95_df2), 
               aes(xintercept = p95), color = colors[6], linetype = "dotted", size = 1.2) +
    
    # Add p999 lines only for non-truncated traces
    geom_vline(data = data.frame(type = df1_name, p999 = p999_df1), 
               aes(xintercept = p999), color = colors[5], linetype = "dashed", size = 1.2) +
    geom_vline(data = data.frame(type = df2_name, p999 = p999_df2), 
               aes(xintercept = p999), color = colors[6], linetype = "dashed", size = 1.2) +
    # Add p100 (max) lines only for non-truncated traces
    geom_vline(data = data.frame(type = df1_name, p100 = p100_df1), 
               aes(xintercept = p100), color = colors[5], linetype = "dotted", size = 1.2) +
    geom_vline(data = data.frame(type = df2_name, p100 = p100_df2), 
               aes(xintercept = p100), color = colors[6], linetype = "dotted", size = 1.2) +
    
    # Add text labels for the lines (vertical orientation)
    geom_text(data = data.frame(type = df1_name_upto95, p50 = p50_df1, y_pos = 0), 
              aes(x = p50, y = y_pos, label = "P50"), 
              color = colors[5], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    geom_text(data = data.frame(type = df1_name_upto95, p95 = p95_df1, y_pos = max_y$ymax[max_y$type == df1_name_upto95]/2), 
              aes(x = p95, y = y_pos, label = "P95"), 
              color = colors[5], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    geom_text(data = data.frame(type = df1_name, p999 = p999_df1, y_pos = 0), 
              aes(x = p999, y = y_pos, label = "P999"), 
              color = colors[5], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    geom_text(data = data.frame(type = df1_name, p100 = p100_df1, y_pos = max_y$ymax[max_y$type == df1_name]/2), 
              aes(x = p100, y = y_pos, label = "P100"), 
              color = colors[5], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    
    geom_text(data = data.frame(type = df2_name_upto95, p50 = p50_df2, y_pos = 0), 
              aes(x = p50, y = y_pos, label = "P50"), 
              color = colors[6], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    geom_text(data = data.frame(type = df2_name_upto95, p95 = p95_df2, y_pos = max_y$ymax[max_y$type == df2_name_upto95]/2), 
              aes(x = p95, y = y_pos, label = "P95"), 
              color = colors[6], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    geom_text(data = data.frame(type = df2_name, p999 = p999_df2, y_pos = 0), 
              aes(x = p999, y = y_pos, label = "P999"), 
              color = colors[6], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    geom_text(data = data.frame(type = df2_name, p100 = p100_df2, y_pos = max_y$ymax[max_y$type == df2_name]/2), 
              aes(x = p100, y = y_pos, label = "P100"), 
              color = colors[6], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    
    labs(title = title,
         x = "Duration",
         y = "Frequency",
         caption = "Solid lines: P50 (Median values) | Dashed lines: P999 | Dotted lines: P100 (Maximum values)") +
    
    theme_minimal() +
    theme(legend.position = "none",  # Remove legend since facet titles show the info
          strip.text = element_text(size = 10, face = "bold"),
          plot.caption = element_text(size = 9, hjust = 0.5))
}

plot_histogram_fortail <- function(df1, df2, df1_name, df2_name, title, colors) {
  p50_df1 = quantile(df1, probs = 0.5, na.rm = TRUE)
  p50_df2 = quantile(df2, probs = 0.5, na.rm = TRUE)
  p95_df1 = quantile(df1, probs = 0.95, na.rm = TRUE)
  p95_df2  = quantile(df2, probs = 0.95, na.rm = TRUE)
  p99_df1 = quantile(df1, probs = 0.999, na.rm = TRUE)
  p99_df2  = quantile(df2, probs = 0.999, na.rm = TRUE)
  p999_df1 = quantile(df1, probs = 0.999, na.rm = TRUE)
  p999_df2  = quantile(df2, probs = 0.999, na.rm = TRUE)
  p100_df1 = quantile(df1, probs = 1.0, na.rm = TRUE)
  p100_df2  = quantile(df2, probs = 1.0, na.rm = TRUE)
  
  df1_name_from95 = paste(df1_name, " from p95", sep="")
  df2_name_from95 = paste(df2_name, " from p95", sep="")
  df1_name_from99 = paste(df1_name, " from p99", sep="")
  df2_name_from99 = paste(df2_name, " from p99", sep="")
  
  # Prepare data with 4 different traces
  data_combined <- data.frame(
    duration = c(df1[df1 >= p95_df1], 
                 df1[df1 >= p99_df1], 
                 df2[df2 >= p95_df2],
                 df2[df2 >= p99_df2]),
    type = factor(c(rep(df1_name_from95, length(df1[df1 >= p95_df1])),
                    rep(df1_name_from99, length(df1[df1 >= p99_df1])),
                    rep(df2_name_from95, length(df2[df2 >= p95_df2])),
                    rep(df2_name_from99, length(df2[df2 >= p99_df2]))),
                  levels = c(df1_name, df1_name_from95, df1_name_from99,
                             df2_name, df2_name_from95, df2_name_from99)
    )
  )
  
  
  max_y <- data_combined %>% group_by(type) %>% summarise(ymax = max(hist(duration, breaks = 30, plot = FALSE)$counts))
  
  # Create color mapping for facets
  colors_mapping <- c(
    df1_name_from95 = colors[1],
    df1_name_from99 = colors[3],
    df2_name_from95 = colors[2],
    df2_name_from99 = colors[4]
  )
  names(colors_mapping) <- c(df1_name_from95, df1_name_from99, df2_name_from95, df2_name_from99)
  
  # Create the plot
  ggplot(data_combined, aes(x = duration, fill = type)) +
    geom_histogram(alpha = 0.8, bins = 30) +
    facet_wrap(~type, scales = "free", nrow = 2, ncol = 2) +
    scale_fill_manual(values = colors_mapping) +
    
    
    # Add p99 lines only for plot p95+ 
    geom_vline(data = data.frame(type = df1_name_from95, p99 = p99_df1), 
               aes(xintercept = p99), color = colors[5], linetype = "dashed", size = 1.2) +
    geom_vline(data = data.frame(type = df2_name_from95, p99 = p99_df2), 
               aes(xintercept = p99), color = colors[6], linetype = "dashed", size = 1.2) +
    # Add p100 lines only for plot p95+ 
    geom_vline(data = data.frame(type = df1_name_from95, p100 = p100_df1), 
               aes(xintercept = p100), color = colors[5], linetype = "dotted", size = 1.2) +
    geom_vline(data = data.frame(type = df2_name_from95, p100 = p100_df2), 
               aes(xintercept = p100), color = colors[6], linetype = "dotted", size = 1.2) +
    
    # Add p999 lines only for plot p99+ 
    geom_vline(data = data.frame(type = df1_name_from99, p999 = p999_df1), 
               aes(xintercept = p999), color = colors[5], linetype = "dashed", size = 1.2) +
    geom_vline(data = data.frame(type = df2_name_from99, p999 = p999_df2), 
               aes(xintercept = p999), color = colors[6], linetype = "dashed", size = 1.2) +
    # Add p100 lines only for plot p99+ 
    geom_vline(data = data.frame(type = df1_name_from99, p100 = p100_df1), 
               aes(xintercept = p100), color = colors[5], linetype = "dotted", size = 1.2) +
    geom_vline(data = data.frame(type = df2_name_from99, p100 = p100_df2), 
               aes(xintercept = p100), color = colors[6], linetype = "dotted", size = 1.2) +

  
    # Add text labels for the lines (vertical orientation)
    geom_text(data = data.frame(type = df1_name_from95, p99 = p99_df1, y_pos = 0), 
              aes(x = p99, y = y_pos, label = "P99"), 
              color = colors[5], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    geom_text(data = data.frame(type = df1_name_from95, p100 = p100_df1, y_pos = max_y$ymax[max_y$type == df1_name_from95]/2), 
              aes(x = p100, y = y_pos, label = "P100"), 
              color = colors[5], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +

    geom_text(data = data.frame(type = df1_name_from99, p999 = p999_df1, y_pos = 0), 
              aes(x = p999, y = y_pos, label = "P999"), 
              color = colors[5], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    geom_text(data = data.frame(type = df1_name_from99, p100 = p100_df1, y_pos = max_y$ymax[max_y$type == df1_name_from99]/2), 
              aes(x = p100, y = y_pos, label = "P100"), 
              color = colors[5], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    
    geom_text(data = data.frame(type = df2_name_from95, p99 = p99_df2, y_pos = 0), 
              aes(x = p99, y = y_pos, label = "P99"), 
              color = colors[6], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    geom_text(data = data.frame(type = df2_name_from95, p100 = p100_df2, y_pos = max_y$ymax[max_y$type == df2_name_from95]/2), 
              aes(x = p100, y = y_pos, label = "P100"), 
              color = colors[6], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    
    geom_text(data = data.frame(type = df2_name_from99, p999 = p999_df2, y_pos = 0), 
              aes(x = p999, y = y_pos, label = "P999"), 
              color = colors[6], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    geom_text(data = data.frame(type = df2_name_from99, p100 = p100_df2, y_pos = max_y$ymax[max_y$type == df2_name_from99]/2), 
              aes(x = p100, y = y_pos, label = "P100"), 
              color = colors[6], vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold", angle = 90) +
    
    labs(title = title,
         x = "Duration",
         y = "Frequency",
         caption = "Solid lines: P50 (Median values) | Dashed lines: P999 | Dotted lines: P100 (Maximum values)") +
    
    theme_minimal() +
    theme(legend.position = "none",  # Remove legend since facet titles show the info
          strip.text = element_text(size = 10, face = "bold"),
          plot.caption = element_text(size = 9, hjust = 0.5))
}

plot_histogram_fullview <- function(df, df_name, title, colors) {
  p50  <- quantile(df, probs = 0.5, na.rm = TRUE)
  p95  <- quantile(df, probs = 0.95, na.rm = TRUE)
  p99  <- quantile(df, probs = 0.99, na.rm = TRUE)
  p999 <- quantile(df, probs = 0.999, na.rm = TRUE)
  p100 <- quantile(df, probs = 1.0, na.rm = TRUE)

  df_name_full    <- paste(df_name, "complete (lines: p999, p100)", sep=" ")
  df_name_upto95  <- paste(df_name, "up to p95 (lines: p50, p95)", sep=" ")
  df_name_from95  <- paste(df_name, "from p95 (lines: p99, p100)", sep=" ")
  df_name_from99  <- paste(df_name, "from p99 (lines: p999, p100)", sep=" ")

  data_combined <- data.frame(
    duration = c(df,
                 df[df <= p95],
                 df[df >= p95],
                 df[df >= p99]),
    type = factor(c(rep(df_name_full,   length(df)),
                    rep(df_name_upto95, length(df[df <= p95])),
                    rep(df_name_from95, length(df[df >= p95])),
                    rep(df_name_from99, length(df[df >= p99]))),
                  levels = c(df_name_full, df_name_upto95, df_name_from95, df_name_from99))
  )

  max_y <- data_combined %>% 
    group_by(type) %>% 
    summarise(ymax = max(hist(duration, breaks = 30, plot = FALSE)$counts), .groups = "drop")

  colors_mapping <- c(
    df_name_full    = colors[1],
    df_name_upto95  = colors[2],
    df_name_from95  = colors[3],
    df_name_from99  = colors[4]
  )
  names(colors_mapping) <- c(df_name_full, df_name_upto95, df_name_from95, df_name_from99)

  lines_df <- data.frame(
    type = factor(c(
      df_name_full, df_name_full,             # Complete
      df_name_upto95, df_name_upto95,         # up to p95
      df_name_from95, df_name_from95,         # from p95
      df_name_from99, df_name_from99          # from p99
    ),
    levels = c(df_name_full, df_name_upto95, df_name_from95, df_name_from99)),
    
    x = c(
      p999, p100,   # complete
      p50, p95,     # up to p95
      p99, p100,    # from p95
      p999, p100    # from p99
    ),
    
    label = c(
      "P999", "P100",
      "P50",  "P95",
      "P99",  "P100",
      "P999", "P100"
    ),
    
    linetype = c(
      "dashed", "dotted",
      "dashed", "dotted",
      "dashed", "dotted",
      "dashed", "dotted"
    ),
    
    color = c(
      colors[5], colors[5],   # complete
      colors[6], colors[6],   # up to p95
      colors[7], colors[7],   # from p95
      colors[8], colors[8]    # from p99
    ),
    
    y_pos = c(
      0, max_y$ymax[max_y$type == df_name_full]/2,
      0, max_y$ymax[max_y$type == df_name_upto95]/2,
      0, max_y$ymax[max_y$type == df_name_from95]/2,
      0, max_y$ymax[max_y$type == df_name_from99]/2
    )
  )
  
  ggplot(data_combined, aes(x = duration, fill = type)) +
    geom_histogram(alpha = 0.8, bins = 30) +
    facet_wrap(~type, scales = "free", nrow = 2, ncol = 2) +  # garante 2x2
    scale_fill_manual(values = colors_mapping) +

    geom_vline(data = lines_df, aes(xintercept = x, linetype = linetype),
               color = lines_df$color, size = 1.2, show.legend = FALSE) +
    geom_text(data = lines_df, aes(x = x, y = y_pos, label = label),
              angle = 90, vjust = 1.5, hjust = -0.1, size = 3, fontface = "bold",
              color = lines_df$color, show.legend = FALSE) +

    labs(title = title,
         x = "Duration",
         y = "Frequency",
         caption = "Dashed lines: P50/P99/P999 | Dotted lines: P95/P100") +
    theme_minimal() +
    theme(legend.position = "none",
          strip.text = element_text(size = 10, face = "bold"),
          plot.caption = element_text(size = 9, hjust = 0.5))
}

call_histograms <- function(df1_latency, df2_latency, df1_name, df2_name, titles, colors) {
  print(plot_histogram(df1_latency, df2_latency, df1_name, df2_name, titles, colors))
  print(plot_histogram_fortail(df1_latency, df2_latency, df1_name, df2_name, titles, colors))
}

calculate_responseTimeToEvaluate <- function(df) {
  df$minResponseTime = pmin(df$responseTime, df$techniqueResponseTime)
  df$responseTimeToEvaluate = ifelse(df$minResponseTime != 0, df$minResponseTime, df$responseTime)
  df$responseTimeToEvaluate = as.numeric(df$responseTimeToEvaluate)
  return(df)
}

summary_table <- function(df1, tag1, df2, tag2, boot_n = 500, conf = 0.95) {
  # Função auxiliar: intervalo de confiança para um quantil via bootstrap
  qCI <- function(df, p) {
    n_boot <- min(length(df), 10000) 
    boots <- replicate(boot_n, {
      sample_df <- sample(df, n_boot, replace = TRUE)
      quantile(sample_df, probs = p, na.rm = TRUE)
    })
    ci <- quantile(boots, probs = c((1 - conf)/2, 1 - (1 - conf)/2), na.rm = TRUE)
    return(ci)
  }
  
  stats <- function(df) {
    # Média com IC (t.test já faz isso)
    avg <- signif(t.test(df)$conf.int, digits = 2)
    
    # Quantis com IC bootstrap
    p50   <- signif(qCI(df, 0.5),   digits = 4)
    p95   <- signif(qCI(df, 0.95),  digits = 4)
    p99   <- signif(qCI(df, 0.99),  digits = 4)
    p999  <- signif(qCI(df, 0.999), digits = 4)
    #dist  <- signif(p999 - p50,     digits = 4)
    
    data <- c(avg, p50, p95, p99, p999)#, dist)
    return(data)
  }
  
  stats1 <- stats(df1)
  stats2 <- stats(df2)
  
  avgdf  <- data.frame("avg",  stats1[1],  stats1[2],  stats2[1],  stats2[2])
  p50df  <- data.frame("p50",  stats1[3],  stats1[4],  stats2[3],  stats2[4])
  p95df  <- data.frame("p95",  stats1[5],  stats1[6],  stats2[5],  stats2[6])
  p99df  <- data.frame("p99",  stats1[7],  stats1[8],  stats2[7],  stats2[8])
  p999df <- data.frame("p999", stats1[9],  stats1[10], stats2[9],  stats2[10])
  #distdf <- data.frame("dist", stats1[11], stats1[12], stats2[11], stats2[12])
  
  tag1_inf <- paste(tag1, "cii", sep = ".")
  tag1_sup <- paste(tag1, "cis", sep = ".")
  tag2_inf <- paste(tag2, "cii", sep = ".")
  tag2_sup <- paste(tag2, "cis", sep = ".")
  
  names(avgdf)  <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(p50df)  <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(p95df)  <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(p99df)  <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  names(p999df) <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  #names(distdf) <- c("stats", tag1_inf, tag1_sup, tag2_inf, tag2_sup)
  
  df <- rbind(avgdf, p50df, p95df, p99df, p999df)#, distdf)
  return(df)
}

plot_replicas_alive <- function(df, snames, scenario_col = "scenario", title = "Replicas Active") {
  df$time_readable <- as.POSIXct(df$timestamp, origin = "2021-01-31", tz = "UTC")
  df[[scenario_col]] <- factor(
    df[[scenario_col]],
    levels = snames
  )
  
  ggplot(df, aes(x = time_readable, y = replica_amount, color = !!sym(scenario_col))) +
    geom_line(size = 0.5) +
    geom_point(size = 0.5) +
    scale_color_manual(values = setNames(
      c("limegreen", "orange", "blue", "red", "purple", "cyan3"),
      snames[1:6]
    )) +
    scale_fill_manual(values=hcl(100, 65, alpha=c(1, 1, 1, 1))) +
    scale_linetype_manual(values = setNames(
      c("solid", "dotted", "dotdash", "dashed", "longdash", "twodash"),
      snames[1:6]
    )) +
    scale_x_datetime(
      date_labels = "%d %b\n%H:%M",   # e.g. "01 Jan\n12:00"
      date_breaks = "2 days"          # tick every 2 days
    ) +
    labs(
      title = title,
      x = "Time (2-week window)",
      y = "Amount of Alive Replicas",
      color = "scenario"
    ) +
    theme_minimal(base_size = 14) +
    theme(
      plot.title = element_text(hjust = 0.5, face = "bold"),
      legend.position = "top"
    )
}

plot_ecdf <- function(title, df, xinf, xsup, yinf, ysup, linetype=c("solid", "dotted", "dotdash", "dashed"), cols=c(15, 195, 150), img_name=FALSE) {
  vanilla.color <- "limegreen"
  vanilla.p95 <- quantile(df$vanilla, 0.95)
  vanilla.p999 <- quantile(df$vanilla, 0.999)
  vanilla.p50 <- quantile(df$vanilla, 0.5)
  
  rh.color <- "purple"
  rh.p95 <- quantile(df$rh, 0.95)
  rh.p999 <- quantile(df$rh, 0.999)
  rh.p50 <- quantile(df$rh, 0.5)
  
  rho.color <- "blue"
  rho.p95 <- quantile(df$rho, 0.95)
  rho.p999 <- quantile(df$rho, 0.999)
  rho.p50 <- quantile(df$rho, 0.5)
  
  gci.color <- "red"
  gci.p95 <- quantile(df$gci, 0.95)
  gci.p999 <- quantile(df$gci, 0.999)
  gci.p50 <- quantile(df$gci, 0.5)
  
  # Definindo alturas diferentes para cada grupo
  vanilla_y = 0.9
  rh_y = 0.1
  rho_y = 0.7
  gci_y = 0.4
  
  size = 0.5
  alpha = 0.8
  angle = 90
  
  p <- df[, colSums(is.na(df)) != nrow(df)] %>%
    pivot_longer(everything()) %>%
    group_by(name) %>%
    arrange(value, by_group = TRUE) %>%
    mutate(ecdf = seq(1/n(), 1 - 1/n(), length.out = n())) %>%
    ggplot(aes(x = value, y = ecdf, colour=name, linetype=name)) +
    xlim(xinf, xsup) +
    ylim(yinf, ysup) +
    geom_step() +
    theme(text=element_text(size=12), plot.title = element_text(hjust = 0.5))+
    theme(legend.text=element_text(size=12)) +
    theme(legend.position="top") +
    
    scale_color_manual(values=c("vanilla"="limegreen", "rh"="purple", "rho"="blue", "gci"="red")) +
    scale_fill_manual(values=hcl(100, 65, alpha=c(1, 1, 1, 1))) +
    scale_linetype_manual(values=c("vanilla"="solid", "rh"="dotted", "rho"="dotdash", "gci"="dashed")) +
    labs(x="Response Time (ms)",y="ECDF", color="scenario", linetype="scenario") +
    ggtitle(title) +
    
    # Vanilla annotations - todos na mesma altura do grupo vanilla
    annotate(geom="text", x=vanilla.p50, y=vanilla_y, label="Median", angle=angle, color=vanilla.color) +
    geom_vline(xintercept=vanilla.p50, linetype="solid", size=size, alpha=alpha, color=vanilla.color) +
    annotate(geom="text", x=vanilla.p95, y=vanilla_y, label="95th", angle=angle, color=vanilla.color) +
    geom_vline(xintercept=vanilla.p95, linetype="solid", size=size, alpha=alpha, color=vanilla.color) +
    annotate(geom="text", x=vanilla.p999, y=vanilla_y, label="99.9th", angle=angle, color=vanilla.color) +
    geom_vline(xintercept=vanilla.p999, linetype="solid", size=size, alpha=alpha, color=vanilla.color) +
    
    # RH annotations - todos na mesma altura do grupo rh
    annotate(geom="text", x=rh.p50, y=rh_y, label="Median", angle=angle, color=rh.color) +
    geom_vline(xintercept=rh.p50, linetype="dotted", size=size, alpha=alpha, color=rh.color) +
    annotate(geom="text", x=rh.p95, y=rh_y, label="95th", angle=angle, color=rh.color) +
    geom_vline(xintercept=rh.p95, linetype="dotted", size=size, alpha=alpha, color=rh.color) +
    annotate(geom="text", x=rh.p999, y=rh_y, label="99.9th", angle=angle, color=rh.color) +
    geom_vline(xintercept=rh.p999, linetype="dotted", size=size, alpha=alpha, color=rh.color) +
    
    # RHO annotations - todos na mesma altura do grupo rho
    annotate(geom="text", x=rho.p50, y=rho_y, label="Median", angle=angle, color=rho.color) +
    geom_vline(xintercept=rho.p50, linetype="dotdash", size=size, alpha=alpha, color=rho.color) +
    annotate(geom="text", x=rho.p95, y=rho_y, label="95th", angle=angle, color=rho.color) +
    geom_vline(xintercept=rho.p95, linetype="dotdash", size=size, alpha=alpha, color=rho.color) +
    annotate(geom="text", x=rho.p999, y=rho_y, label="99.9th", angle=angle, color=rho.color) +
    geom_vline(xintercept=rho.p999, linetype="dotdash", size=size, alpha=alpha, color=rho.color) +
    
    # GCI annotations - todos na mesma altura do grupo gci
    annotate(geom="text", x=gci.p50, y=gci_y, label="Median", angle=angle, color=gci.color) +
    geom_vline(xintercept=gci.p50, linetype="dashed", size=size, alpha=alpha, color=gci.color) +
    annotate(geom="text", x=gci.p95, y=gci_y, label="95th", angle=angle, color=gci.color) +
    geom_vline(xintercept=gci.p95, linetype="dashed", size=size, alpha=alpha, color=gci.color) +
    annotate(geom="text", x=gci.p999, y=gci_y, label="99.9th", angle=angle, color=gci.color) +
    geom_vline(xintercept=gci.p999, linetype="dashed", size=size, alpha=alpha, color=gci.color) +
    
    theme_bw()
  
  if (img_name != FALSE) {
    ggsave(img_name, width=10, height=5)
  }
  print(p)
}

load_data_summarized <- function(results_path, techs, probs, idletimes) {
  parse_idletime_and_probs <- function(df) {
    df$tailLatencyProb = ifelse(df$tailLatencyProb == "tlprobp95", 95, df$tailLatencyProb)
    df$tailLatencyProb = ifelse(df$tailLatencyProb == "tlprobp99", 99, df$tailLatencyProb)
    
    df$idletime = ifelse(df$idletime == "idletime0.0", 0, df$idletime)
    df$idletime = ifelse(df$idletime == "idletime900.0", 900, df$idletime)
    df$idletime = ifelse(df$idletime == "idletimeINF", Inf, df$idletime)
    
    return(df)
  }
  
  all_invocs <- list()
  all_replicas <- list()
  for (t in techs) {
    for (p in probs) {
      for (i in idletimes) {
        scenario = paste(t, i, p, sep = "_")
        fileName = paste(scenario, "-invocations.csv", sep="")
        df_invocs = read.csv(paste(results_path, t, fileName, sep="/"), header = TRUE)
        df_invocs$minResponseTime = pmin(df_invocs$responseTime, df_invocs$techniqueResponseTime)
        df_invocs$responseTimeToEvaluate = ifelse(
          df_invocs$minResponseTime != 0,
          df_invocs$minResponseTime,
          df_invocs$responseTime
        )
        df_invocs$technique = t
        df_invocs$idletime = i
        df_invocs$tailLatencyProb = p
        
        df_invocs <- df_invocs %>%
          group_by(funcID) %>%
          mutate(
            p50  = quantile(responseTimeToEvaluate, probs = 0.50, na.rm = TRUE),
            p95  = quantile(responseTimeToEvaluate, probs = 0.95, na.rm = TRUE),
            p99  = quantile(responseTimeToEvaluate, probs = 0.99, na.rm = TRUE),
            p999 = quantile(responseTimeToEvaluate, probs = 0.999, na.rm = TRUE),
            p100 = quantile(responseTimeToEvaluate, probs = 1.00, na.rm = TRUE)
          ) %>%
          ungroup()
        
        df_invocs <- df_invocs %>% distinct(funcID, .keep_all = TRUE)
        all_invocs[[length(all_invocs) + 1]] <- data.frame(
          funcID          = df_invocs$funcID,
          technique       = df_invocs$technique,
          tailLatencyProb = df_invocs$tailLatencyProb,
          idletime        = df_invocs$idletime,
          p50             = df_invocs$p50,
          p95             = df_invocs$p95,
          p99             = df_invocs$p99,
          p999            = df_invocs$p999,
          p100            = df_invocs$p100
        )
        
        fileName = paste(scenario, "-replicas.csv", sep="")
        df_replicas = read.csv(paste(results_path, t, fileName, sep="/"), header = TRUE)
        df_replicas$technique = t
        df_replicas$idletime = i
        df_replicas$tailLatencyProb = p
        df_replicas$replicasCount = nrow(filter(df_replicas, funcID == df_replicas$funcID))
        df_replicas <- df_replicas %>%
          group_by(funcID) %>%
          mutate(
            totalUptime = sum(df_replicas$upTime) /1000/60/60,
            totalBusyTime = sum(df_replicas$busyTime) /1000/60/60,
          ) %>%
          ungroup()
        
        
        df_replicas <- df_replicas %>% distinct(funcID, .keep_all = TRUE)
        all_replicas[[length(all_replicas) + 1]] <- data.frame(
          funcID          = df_replicas$funcID,
          technique       = df_replicas$technique,
          tailLatencyProb = df_replicas$tailLatencyProb,
          idletime        = df_replicas$idletime,
          totalUptime     = df_replicas$totalUptime,
          totalBusyTime   = df_replicas$totalBusyTime,
          replicasCount   = df_replicas$replicasCount
        )
      }
    }
  }
  df_for_invocs <- parse_idletime_and_probs(do.call(rbind, all_invocs))
  df_for_replicas <- parse_idletime_and_probs(do.call(rbind, all_replicas))
  return(list(invocations_df = df_for_invocs, replicas_df = df_for_replicas))
}

plot_hist_summarized <- function(df, metric, y_label, title) {
  df$idletime_f <- factor(df$idletime, levels = c(0, 900, Inf))
  ggplot(df, aes(x=tailLatencyProb, y=metric, color=as.factor(technique), shape=as.factor(technique))) +
    geom_jitter(width = 0.015, height = 0, alpha = 0.8, size = 2.5, stroke = 1) +
    facet_grid(. ~ factor(idletime_f, levels = c(0, 900, Inf))) +
    labs(x = "tail latency probability", y = y_label, title = title, colour = "technique", shape = "technique") +
    scale_shape_manual(values = c(1, 23, 24, 15)) +
    scale_color_manual(values = c("red", "purple", "blue", "limegreen")) +  # cores mais distintas
    theme_bw() +
    theme(
      legend.position = "bottom",
      strip.text = element_text(size = 12),
      axis.text = element_text(size = 10)
    )
}
